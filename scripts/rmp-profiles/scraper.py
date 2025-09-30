from seleniumwire import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from selenium.common.exceptions import TimeoutException
import re
import time
import json
import datetime
import requests


def setup_driver(headless=True):
    """Sets up and returns a Selenium WebDriver."""
    try:
        chrome_options = Options()
        chrome_options.add_argument("--log-level=3")
        chrome_options.add_argument("--ignore-certificate-errors")
        if headless:
            chrome_options.add_argument("--headless")
        driver = webdriver.Chrome(options=chrome_options)
        driver.set_page_load_timeout(5)
        return driver
    except Exception as e:
        print(f"Failed to start the Chrome driver: {e}")
        print("Go to https://googlechromelabs.github.io/chrome-for-testing/#stable to download the latest version of ChromeDriver. Copy the executable to the root folder of this project. You may also need the latest version of Chrome; make sure your chrome is updated.")
        exit(1)


def close_cookie_popup(driver):
    """Closes the cookie popup if it exists."""
    try:
        WebDriverWait(driver, 5).until(
            EC.presence_of_element_located((By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2"))
        )
        close_button = driver.find_element(By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2")
        driver.execute_script("arguments[0].click();", close_button)
        print("Cookie popup closed.")
        time.sleep(2)
    except Exception as e:
        print("No cookie popup found or issue clicking it:", e)

def click_pagination_button(driver):
    """Clicks the pagination button to load more professors."""
    try:
        pagination_button = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.CLASS_NAME, "PaginationButton__StyledPaginationButton-txi1dr-1"))
        )
        driver.execute_script("arguments[0].scrollIntoView(true);", pagination_button)
        driver.execute_script("arguments[0].click();", pagination_button)
        print("Clicked on the pagination button.")
        time.sleep(5)
    except Exception as e:
        print(f"Failed to find or click the pagination button: {e}")


def get_headers(driver, school_id):
    """Gets the necessary headers and school ID from the GraphQL request."""
    url = f'https://www.ratemyprofessors.com/search/professors/{school_id}?q=*'

    # go to the rmp page for UTD
    try:
        driver.get(url)
    except TimeoutException:
        driver.execute_script("window.stop();")
        try:
            driver.refresh()
        except TimeoutException:
            driver.execute_script("window.stop();")
        time.sleep(2)

    # close cookie popup if it exists
    close_cookie_popup(driver)

    # click on the "show more professors" button to trigger the graphql request
    click_pagination_button(driver)

    # find graphql headers from the request
    url_filter = "ratemyprofessors.com/graphql"
    graphql_headers = {}
    for request in driver.requests:
        if request.response and url_filter in request.url:
            print(f"\n[REQUEST] {request.url}")
            request_body = request.body
            m = re.findall(r'schoolID":"(.*?)"', str(request_body))
            if m:
                print(f"\tschoolID: {m[0]}")
            else:
                print("schoolID not found in request body.")
                return None, None
            print("Headers:")
            graphql_headers = request.headers
            for header, value in request.headers.items():
                print(f"\t{header}: {value}")
            print("-" * 50)
            return graphql_headers, m[0]
    return None, None


def normalize_course_name(course_name):
    """Normalizes a course name to uppercase and removes spaces and hyphens."""
    return re.sub(r'[-_\s]+', '', course_name).upper()


def normalize_professor_name(name):
    """Normalizes professor names by removing extra spaces and converting to lowercase."""
    return " ".join(name.lower().split())


def build_graphql_query():
    """Builds the GraphQL query for retrieving professor data."""
    return """query TeacherSearchPaginationQuery( $count: Int!  $cursor: String $query: TeacherSearchQuery!) { search: newSearch { ...TeacherSearchPagination_search_1jWD3d } }
        fragment TeacherSearchPagination_search_1jWD3d on newSearch {
            teachers(query: $query, first: $count, after: $cursor) {
                didFallback
                edges {
                    cursor
                    node {
                        ...TeacherCard_teacher
                        id
                        __typename
                    }
                }
                pageInfo {
                    hasNextPage
                    endCursor
                }
                resultCount
                filters {
                    field
                    options {
                        value
                        id
                    }
                }
            }
        }
        fragment TeacherCard_teacher on Teacher {
            id
            legacyId
            avgRating
            numRatings
            courseCodes {
                courseName
                courseCount
            }
            ...CardFeedback_teacher
            ...CardSchool_teacher
            ...CardName_teacher
            ...TeacherBookmark_teacher
            ...TeacherTags_teacher
        }
        fragment CardFeedback_teacher on Teacher {
            wouldTakeAgainPercent
            avgDifficulty
        }
        fragment CardSchool_teacher on Teacher {
            department
            school {
                name
                id
            }
        }
        fragment CardName_teacher on Teacher {
            firstName
            lastName
        }
        fragment TeacherBookmark_teacher on Teacher {
            id
            isSaved
        }
        fragment TeacherTags_teacher on Teacher {
            lastName
            teacherRatingTags {
                legacyId
                tagCount
                tagName
                id
            }
        }
    """


def transform_professor_data(professor_node):
    """Transforms a single professor node from GraphQL response into our data format."""
    dn = professor_node['node']
    
    # Extract and sort tags
    tags = []
    if dn['teacherRatingTags']:
        sorted_tags = sorted(dn['teacherRatingTags'], key=lambda x: x['tagCount'], reverse=True)
        tags = [tag['tagName'] for tag in sorted_tags[:5]]
    
    # Extract and normalize courses
    courses = [normalize_course_name(course['courseName']) for course in dn['courseCodes']]
    courses = list(set(courses))
    
    # Build profile URL
    profile_link = f"https://www.ratemyprofessors.com/professor/{dn['legacyId']}" if dn['legacyId'] else None
    
    return {
        'department': dn['department'],
        'url': profile_link,
        'quality_rating': dn['avgRating'],
        'difficulty_rating': dn['avgDifficulty'],
        'would_take_again': round(dn['wouldTakeAgainPercent']),
        'original_rmp_format': f"{dn['firstName']} {dn['lastName']}",
        'last_updated': datetime.datetime.now().isoformat(),
        'ratings_count': dn['numRatings'],
        'courses': courses,
        'tags': tags,
        'rmp_id': str(dn['legacyId'])
    }


def execute_graphql_request(headers, req_data, max_retries=3):
    """Executes a single GraphQL request with retry logic."""
    for attempt in range(max_retries):
        try:
            res = requests.post(
                'https://www.ratemyprofessors.com/graphql', 
                headers=headers, 
                json=req_data,
                timeout=30
            )

            if res.status_code != 200:
                print(f"HTTP Error: {res.status_code} on attempt {attempt + 1}")
                if attempt < max_retries - 1:
                    time.sleep(2)
                    continue
                else:
                    print("Failed after all HTTP retries.")
                    return None

            return res.json()['data']['search']['teachers']['edges']
            
        except (json.JSONDecodeError, KeyError, requests.exceptions.RequestException) as e:
            print(f"Error in GraphQL request (attempt {attempt + 1}): {e}")
            if attempt < max_retries - 1:
                print("Retrying GraphQL request...")
                time.sleep(2)
            else:
                print("Failed after all GraphQL retries.")
                return None
    
    return None


def query_rmp(headers, school_id, max_retries=3):
    """Queries the internal RMP GraphQL API to retrieve professor data."""
    # thank you Michael Zhao for this idea
    max_prof_count = 1000  # maximum number of professors per request

    req_data = {
        "query": build_graphql_query(),
        "variables": {
            "count": max_prof_count,
            "cursor": "",
            "query": {
                "text": "",
                "schoolID": school_id,
                "fallback": True
            }
        }
    }

    all_professors = {}
    more = True
    
    while more:
        more = False
        
        data = execute_graphql_request(headers, req_data, max_retries)
        if not data:
            break
            
        # process each professor in the response
        for professor_node in data:
            professor_data = transform_professor_data(professor_node)
            professor_name = f"{professor_node['node']['firstName']} {professor_node['node']['lastName']}"
            key = normalize_professor_name(professor_name)

            if key in all_professors:
                all_professors[key].append(professor_data)
                print(f"Duplicate RMP professor name found: {key}")
            else:
                all_professors[key] = [professor_data]

        # check if there are more pages (only 1000 results max per page, query again if we have the max results)
        if len(data) == max_prof_count:
            req_data['variables']['cursor'] = data[len(data) - 1]['cursor']
            more = True

    return all_professors


def scrape_rmp_data_attempt(university_id):
    start_time = time.time()

    driver = setup_driver()
    setup_driver_time = time.time()
    print(f"Driver setup time: {setup_driver_time - start_time:.2f} seconds")

    try:
        headers, school_id = get_headers(driver, university_id)
        get_headers_time = time.time()
        print(
            f"Get headers time: {get_headers_time - setup_driver_time:.2f} seconds")

        if headers and school_id:
            professor_data = query_rmp(headers, school_id)
            query_rmp_time = time.time()
            print(
                f"Query RMP time: {query_rmp_time - get_headers_time:.2f} seconds")

            if professor_data:
                end_time = time.time()
                print(f"Attempt execution time: {end_time - start_time:.2f} seconds")
                return professor_data
            else:
                print("Data extraction failed. GraphQL API returned no data.")
                return None
        else:
            print("Failed to retrieve headers or school ID. Data extraction aborted.")
            return None
    finally:
        driver.quit()


def scrape_rmp_data(university_id, max_retries=3):
    """Scrapes professor data from RateMyProfessors with retry logic."""
    overall_start_time = time.time()
    
    for attempt in range(max_retries):
        try:
            print(f"\n=== Scraping attempt {attempt + 1}/{max_retries} ===")
            
            result = scrape_rmp_data_attempt(university_id)
            
            if result:
                overall_end_time = time.time()
                print(f"Total execution time (including retries): {overall_end_time - overall_start_time:.2f} seconds")
                return result
            else:
                if attempt < max_retries - 1:
                    print(f"Attempt {attempt + 1} failed. Retrying in 3 seconds...")
                    time.sleep(3)
                else:
                    print(f"All {max_retries} attempts failed.")
                    
        except Exception as e:
            print(f"Exception occurred on attempt {attempt + 1}: {e}")
            if attempt < max_retries - 1:
                print(f"Retrying in 3 seconds...")
                time.sleep(3)
            else:
                print(f"All {max_retries} attempts failed due to exceptions.")
    
    overall_end_time = time.time()
    print(f"Total execution time (including failed retries): {overall_end_time - overall_start_time:.2f} seconds")
    return {}


# old webscrape implementation, it seems the graphql api is much faster but has a tendency to occasionally not return all the data, such as courses, tags, and the would_take_again values, might still need this

# def extract_professor_data(page_source):
#     """Extracts and normalizes professor data from the page source."""
#     try:
#         # duplicate_prof_count = 0
#         print("Parsing data...")
#         soup = BeautifulSoup(page_source, "html.parser")
#         professors = soup.find_all("a", class_="TeacherCard__StyledTeacherCard-syjs0d-0")

#         if not professors:
#             print("No professor data found. The page source might be incomplete.")
#             return {}

#         professor_data = {}

#         for prof in professors:
#             name_tag = prof.find("div", class_="CardName__StyledCardName-sc-1gyrgim-0")
#             rating_tag = prof.find("div", class_="CardNumRating__CardNumRatingNumber-sc-17t4b9u-2")
#             department_tag = prof.find("div", class_="CardSchool__Department-sc-19lmz2k-0")
#             would_take_again_tag = prof.find("div", class_="CardFeedback__CardFeedbackNumber-lq6nix-2")
#             difficulty_tag = prof.find_all("div", class_="CardFeedback__CardFeedbackNumber-lq6nix-2")
#             ratings_count_tag = prof.find("div", class_="CardNumRating__CardNumRatingCount-sc-17t4b9u-3") # Get the ratings count tag

#             name = name_tag.text.strip() if name_tag else "Unknown"
#             department = department_tag.text.strip() if department_tag else "Unknown"

#             rating_text = rating_tag.text.strip() if rating_tag else "N/A"
#             rating = float(rating_text) if rating_text != "N/A" else "N/A"
#             would_take_again_text = would_take_again_tag.text.strip().replace('%', '') if would_take_again_tag else "N/A"
#             would_take_again = float(would_take_again_text) if would_take_again_text != "N/A" else "N/A"
#             difficulty_text = difficulty_tag[1].text.strip() if difficulty_tag and len(difficulty_tag) > 1 else "N/A"
#             difficulty = float(difficulty_text) if difficulty_text != "N/A" else "N/A"
#             ratings_count_text = ratings_count_tag.text.strip().replace(" ratings", "") if ratings_count_tag else "N/A"
#             ratings_count = int(ratings_count_text) if ratings_count_text != "N/A" else "N/A"

#             prof_url = "https://www.ratemyprofessors.com" + prof['href']
#             prof_id = prof['href'].split('/')[-1]

#             normalized_name = " ".join(name.lower().split())

#             # if(normalized_name in professor_data):
#             #     duplicate_prof_count+=1
#             #     print(f"Duplicate professor found: {normalized_name}, counter = {duplicate_prof_count}")

#             professor_data[normalized_name] = {
#                 "id": prof_id,
#                 "department": department,
#                 "url": prof_url,
#                 "quality_rating": rating,
#                 "difficulty_rating": difficulty,
#                 "would_take_again": would_take_again,
#                 "original_format": name,
#                 "last_updated": datetime.datetime.now().isoformat(),
#                 "ratings_count": ratings_count
#             }
#         return professor_data
#     except Exception as e:
#         print(f"Error extracting professor data: {e}")
#         return {}

# async def fetch_professor_data_async(session, url):
#     """Fetches and parses a professor's reviews and tags from RMP."""
#     try:
#         async with session.get(url) as response:
#             response.raise_for_status()
#             html = await response.text()
#             soup = BeautifulSoup(html, 'html.parser')

#             # store courses and tags as sets to avoid duplicates but convert to lists for JSON serialization
#             course_tags = soup.find_all("div", class_="RatingHeader__StyledClass-sc-1dlkqw1-3")
#             courses = {normalize_course_name(tag.text.strip()) for tag in course_tags} # only 20 ratings displayed on the page but we can hope, also normalize the course names

#             tag_tags = soup.find_all("span", class_="Tag-bs9vf4-0")
#             tags = {tag.text.strip() for tag in tag_tags}

#             return {"courses": list(courses), "tags": list(tags)[:5]}

#     except aiohttp.ClientError as e:
#         print(f"Error fetching {url}: {e}")
#         return {"courses": [], "tags": []}
#     except Exception as e:
#         print(f"Error parsing {url}: {e}")
#         return {"courses": [], "tags": []}

# async def scrape_professor_courses_async(professor_data):
#     """Scrapes course reviews and tags for all professors in the provided data."""
#     urls = [prof_data["url"] for prof_data in professor_data.values() if "url" in prof_data]

#     start_time = time.time()

#     async with aiohttp.ClientSession() as session:
#         tasks = [fetch_professor_data_async(session, url) for url in urls]
#         results = await asyncio.gather(*tasks)

#     for i, data in enumerate(results):
#         professor_name = list(professor_data.keys())[i]
#         if professor_name in professor_data:
#             professor_data[professor_name]["courses"] = data["courses"]
#             professor_data[professor_name]["tags"] = data["tags"]

#     end_time = time.time()
#     print(f"Scraped course data and tags for {len(professor_data)} professors in {end_time - start_time:.2f} seconds.")

#     return professor_data
