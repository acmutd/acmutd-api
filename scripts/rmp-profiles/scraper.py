from seleniumwire import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException
import re
import time
import json
import datetime
import requests

def setup_driver(headless=True):
    """Sets up and returns a Selenium WebDriver."""
    driver = None
    try:
        chrome_options = Options()
        chrome_options.add_argument("--log-level=3")
        chrome_options.add_argument("--ignore-certificate-errors")
        if headless:
            chrome_options.add_argument("--headless")
        driver = webdriver.Chrome(options=chrome_options)
        driver.set_page_load_timeout(2)
        return driver
    except Exception as e:
        print(f"Failed to start the Chrome driver: {e}")
        print("Go to https://googlechromelabs.github.io/chrome-for-testing/#stable to download the latest version of ChromeDriver. Copy the executable to the root folder of this project. You may also need the latest version of Chrome; make sure your chrome is updated.")
        exit(1)


def close_cookie_popup(driver):
    """Closes the cookie popup if it exists."""
    try:
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2"))
        )
        close_button = driver.find_element(By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2")
        driver.execute_script("arguments[0].click();", close_button)
        print("Cookie popup closed.")
        time.sleep(2)
    except Exception as e:
        print("No cookie popup found or issue clicking it:", e)


def get_headers(driver, school_id):
    """Gets the necessary headers and school ID from the GraphQL request."""
    try:
        driver.get(f'https://www.ratemyprofessors.com/search/professors/{school_id}?q=*')
    except TimeoutException:
        driver.execute_script("window.stop();")
        try:
            driver.refresh()
        except TimeoutException:
            driver.execute_script("window.stop();")
        time.sleep(2)

    try:
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2"))
        )
        close_button = driver.find_element(By.CLASS_NAME, "CCPAModal__StyledCloseButton-sc-10x9kq-2")
        driver.execute_script("arguments[0].click();", close_button)
        print("Cookie popup closed.")
        time.sleep(2)
    except Exception as e:
        print("No cookie popup found or issue clicking it:", e)

    try:
        pagination_button = WebDriverWait(driver, 5).until(
            EC.element_to_be_clickable((By.CLASS_NAME, "PaginationButton__StyledPaginationButton-txi1dr-1"))
        )
        driver.execute_script("arguments[0].click();", pagination_button)
        print("Clicked on the pagination button.")
    except Exception as e:
        print(f"Failed to find or click the pagination button: {e}")

    time.sleep(4)

    url_filter = "ratemyprofessors.com/graphql"
    graphql_headers = {}
    for request in driver.requests:
        if url_filter in request.url:
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


def query_rmp(headers, school_id):
    """Queries the RMP GraphQL API to retrieve professor data."""
    # thank you Michael Zhao for this idea
    req_data = {
        "query": """query TeacherSearchPaginationQuery( $count: Int!  $cursor: String $query: TeacherSearchQuery!) { search: newSearch { ...TeacherSearchPagination_search_1jWD3d } }
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
        """,
        "variables": {
            "count": 1000,
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
        res = requests.post('https://www.ratemyprofessors.com/graphql', headers=headers, json=req_data)

        if res.status_code != 200:
            print(f"HTTP Error: {res.status_code}. Aborting.")
            return all_professors

        try:
            data = res.json()['data']['search']['teachers']['edges']
            for d in data:
                dn = d['node']

                tags = []
                if dn['teacherRatingTags']:
                    sorted_tags = sorted(dn['teacherRatingTags'], key=lambda x: x['tagCount'], reverse=True)
                    tags = [tag['tagName'] for tag in sorted_tags[:5]]

                courses = [normalize_course_name(course['courseName']) for course in dn['courseCodes']]
                courses = list(set(courses))

                profile_link = f"https://www.ratemyprofessors.com/professor/{dn['legacyId']}" if dn['legacyId'] else None

                professor_data = {
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

                professor_name = f"{dn['firstName']} {dn['lastName']}"
                key = normalize_professor_name(professor_name)

                if key in all_professors:
                    all_professors[key].append(professor_data)
                    print(f"Duplicate RMP professor name found: {key}")
                else:
                    all_professors[key] = [professor_data]

            if len(data) == 1000:
                req_data['variables']['cursor'] = data[len(data) - 1]['cursor']
                more = True
        except json.JSONDecodeError:
            print("Invalid JSON response. Aborting.")
            return all_professors
        except KeyError as e:
            print(f"Missing Key in JSON: {e}. Aborting.")
            return all_professors

    return all_professors


def scrape_rmp_data(university_id):
    """Scrapes professor data from RateMyProfessors."""
    start_time = time.time()  # Start time tracking

    driver = setup_driver()
    setup_driver_time = time.time()
    print(f"Driver setup time: {setup_driver_time - start_time:.2f} seconds")

    headers, school_id = get_headers(driver, university_id)
    get_headers_time = time.time()
    print(f"Get headers time: {get_headers_time - setup_driver_time:.2f} seconds")

    driver.quit()

    if headers and school_id:
        professor_data = query_rmp(headers, school_id)
        query_rmp_time = time.time()
        print(f"Query RMP time: {query_rmp_time - get_headers_time:.2f} seconds")

        if professor_data:
            with open("ratings/rmp_ratings.json", "w", encoding="utf-8") as f:
                json.dump(professor_data, f, indent=4, ensure_ascii=False)
            print("Data extraction and file writing complete.")
            end_time = time.time()
            print(f"Total execution time: {end_time - start_time:.2f} seconds")
            return professor_data
        else:
            print("Data extraction failed. GraphQL API returned no data.")
            end_time = time.time()
            print(f"Total execution time: {end_time - start_time:.2f} seconds")
            return None
    else:
        print("Failed to retrieve headers or school ID. Data extraction aborted.")
        end_time = time.time()
        print(f"Total execution time: {end_time - start_time:.2f} seconds")
        return None




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