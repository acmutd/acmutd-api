import requests
import re
import json
import os
from bs4 import BeautifulSoup
from login import get_cookie

base_url = 'https://coursebook.utdallas.edu'
url = 'https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog'
output = 'classes.json'

DROPDOWN_PREFIX_ID = 'combobox_cp'
DROPDOWN_SCHOOL_ID = 'combobox_col'
DROPDOWN_DAYS_ID = 'combobox_days'
DROPDOWN_LEVELS_ID = 'combobox_clevel'
DROPDOWN_TERM_ID = 'combobox_term'

FILTER_TYPES_MAP = {
    'prefix': DROPDOWN_PREFIX_ID,
    'school': DROPDOWN_SCHOOL_ID,
    'day': DROPDOWN_DAYS_ID,
    'level': DROPDOWN_LEVELS_ID,
}


def get_latest_term():
    try:
        res = requests.get(base_url, timeout=5)
        res.raise_for_status()
    except requests.exceptions.RequestException as e:
        print(f'Failed to get coursebook website: {e}')
        return {}

    pattern = fr'<select .*?id="{re.escape(DROPDOWN_TERM_ID)}".*?>\s*(.*?)\s*</select>'
    matches = re.findall(pattern, res.text, re.DOTALL)
    if not matches:
        print(f"Warning: Failed to find dropdown with ID '{DROPDOWN_TERM_ID}'")
        return {}

    raw_options = matches[0]
    values = re.findall(r'value="([^"]+)"', raw_options)

    latest_term = values[2]
    return latest_term.split('_')[1]


def get_dropdown_options(dropdown_ids):
    try:
        res = requests.get(base_url, timeout=5)
        res.raise_for_status()
    except requests.exceptions.RequestException as e:
        print(f'Failed to get coursebook website: {e}')
        return {}

    options_data = {}

    # for each dropdown id, match the <select> element and extract the options
    for dropdown_id in dropdown_ids:
        pattern = fr'<select .*?id="{re.escape(dropdown_id)}".*?>\s*(.*?)\s*</select>'
        matches = re.findall(pattern, res.text, re.DOTALL)

        if not matches:
            print(f"Warning: Failed to find dropdown with ID '{dropdown_id}'")
            options_data[dropdown_id] = []
            continue

        raw_options = matches[0]
        values = re.findall(r'value="([^"]+)"', raw_options)

        # filter out empty and "any" values, i.e. "Any School" or "Any Level"
        values = [v for v in values if v.strip(
        ) and not v.lower().startswith("any")]

        options_data[dropdown_id] = values

    return options_data


def make_course_request(session_id, term, prefix=None, school=None, day=None, level=None):
    """
    Perform a POST to coursebook with dynamically added filters.
    """
    headers = {
        'accept': '*/*',
        'accept-language': 'en-US,en;q=0.9',
        'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
        'cookie': f'PTGSESSID={session_id}',
        'origin': 'https://coursebook.utdallas.edu',
        'priority': 'u=1, i',
        'referer': 'https://coursebook.utdallas.edu/',
        'sec-ch-ua': '"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"Linux"',
        'sec-fetch-dest': 'empty',
        'sec-fetch-mode': 'cors',
        'sec-fetch-site': 'same-origin',
        'user-agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36',
        'x-requested-with': 'XMLHttpRequest',
    }

    # dynamically build the list of filters to include in the request
    s_params = [f'term_{term}']
    if prefix:
        s_params.append(prefix)
    if school:
        s_params.append(school)
    if day:
        s_params.append(day)
    if level:
        s_params.append(level)

    data = {
        'action': 'search',
        's[]': s_params
    }

    response = requests.post(url, headers=headers, data=data, timeout=12)
    if response.status_code != 200:
        raise Exception(f"Failed course request: {response.text[:200]}")

    return response


def make_monkey_request(session_id, report_id):
    monkey_headers = {
        'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
        'accept-language': 'en-US,en;q=0.9',
        'cookie': f'PTGSESSID={session_id}',
        'priority': 'u=0, i',
        'referer': 'https://coursebook.utdallas.edu/',
        'sec-ch-ua': '"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"Linux"',
        'sec-fetch-dest': 'document',
        'sec-fetch-mode': 'navigate',
        'sec-fetch-site': 'same-origin',
        'sec-fetch-user': '?1',
        'upgrade-insecure-requests': '1',
        'user-agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36',
    }

    monkey_url = f'https://coursebook.utdallas.edu/reportmonkey/cb11-export/{report_id}/json'
    monkey_response = requests.get(monkey_url, headers=monkey_headers)
    return monkey_response


# Get extra class overview detail for waitlist)
def get_class_detail(session_id, section_address, data_req, div_id):
    url = "https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog"

    headers = {
        'accept': '*/*',
        'accept-language': 'en-US,en;q=0.9',
        'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
        'cookie': f'PTGSESSID={session_id}',
        'origin': 'https://coursebook.utdallas.edu',
        'priority': 'u=1, i',
        'referer': 'https://coursebook.utdallas.edu/guidedsearch',
        'sec-ch-ua': '"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"Linux"',
        'sec-fetch-dest': 'empty',
        'sec-fetch-mode': 'cors',
        'sec-fetch-site': 'same-origin',
        'user-agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36',
        'x-requested-with': 'XMLHttpRequest',
    }
    
    data = {
        "id": section_address,
        "req": data_req,
        "action": "info",
        "div": div_id
    }

    response = requests.post(url, headers=headers, data=data, timeout=12)

    return response.text


def parse_class_overview(html, section_addr):
    soup = BeautifulSoup(html, "html.parser")
    
    # Some sections are separated by th title then td content
    def get_val(label):
        """Finds label in header cells and returns the next td's text"""
        for th in soup.find_all('th'):
            if re.search(re.escape(label), th.get_text(), re.I):
                val_cell = th.find_next('td')
                return val_cell.get_text(strip=True) if val_cell else None
        for td in soup.find_all('td', class_='courseinfo__classsubtable__th'):
            if re.search(re.escape(label), td.get_text(), re.I):
                val_cell = td.find_next_sibling('td')
                return val_cell.get_text(strip=True) if val_cell else None
        return None

    def get_list_val(label):
        """Finds label in header cells and returns a list of <li> text values from the next td."""
        for th in soup.find_all('th'):
            if re.search(re.escape(label), th.get_text(), re.I):
                val_cell = th.find_next('td')
                if val_cell:
                    return [li.get_text(strip=True) for li in val_cell.find_all('li')]
        return []

    def get_school():
        for th in soup.find_all('th'):
            if re.search(r'College', th.get_text(), re.I):
                td = th.find_next('td')
                if td:
                    name = td.get_text(strip=True)
                    a = td.find('a', href=True)
                    school_id = None
                    if a:
                        match = re.search(r'https?://(\w+)\.utdallas', a['href'])
                        school_id = match.group(1) if match else None
                    return name, school_id
        return None, None
    
    def get_enrolled_status(status_text):
        if not status_text:
            return 'CLOSED'
        if 'CANCELLED' in status_text:
            return 'CANCELLED'
        if 'OPEN' in status_text:
            return 'OPEN'
        return 'CLOSED'


    def get_syllabus():
        for th in soup.find_all('th'):
            if re.search(r'Syllabus', th.get_text(), re.I):
                td = th.find_next('td')
                if td:
                    a = td.find('a', href=True)
                    if a:
                        match = re.search(r'/(syl\w+)', a['href'])
                        return match.group(1) if match else None
        return None
    
    
    def parse_people():
        """Parse Instructors and TAs"""
        people = {
            'instructors': [], 'instructor_ids': [],
            'tas': [], 'ta_ids': []
        }
        
        # Iterate all divs with id starting with 'inst-'
        for div in soup.select("div[id^='inst-']"):
            text_content = div.get_text(separator="・", strip=True)
            name = text_content.split("・")[0].strip()
            
            # Parse netid from mailto link
            email_link = div.find("a", href=re.compile("mailto:"))
            net_id = email_link['href'].replace('mailto:', '').split('@')[0] if email_link else ""
            
            if "Teaching Assistant" in text_content:
                people['tas'].append(name)
                if net_id: people['ta_ids'].append(net_id)
            else:
                people['instructors'].append(name)
                if net_id: people['instructor_ids'].append(net_id)
                
        return people


    def parse_location_and_schedule():
        """Gets location code and schedule details"""
        meeting_div = soup.find('div', class_='courseinfo__meeting-item--multiple')
        if not meeting_div:
            if status_row_text and 'CANCELLED' in status_row_text:
                print(f"Cancelled class (no meeting): {section_addr}")
            else:
                print(f"Warning: meeting div not found for {section_addr}")
            return {'days': [], 'times_12h': "", 'location': ""}

        lines = list(meeting_div.stripped_strings)

        days_raw = lines[1] if len(lines) > 1 else None                                                                                                                                                                                        
        days = [d.strip() for d in days_raw.split(',')] if days_raw else []   
        
        loc_link = meeting_div.find('a', href=re.compile(r"locator\.utdallas\.edu"))

        if loc_link:
            location = loc_link.get_text(strip=True)
        else:
            # Fallback for when there is no link ("See instructor for room assignment")
            map_div = meeting_div.find('div', class_='courseinfo__map')
            if map_div and map_div.get_text(strip=True):
                location = map_div.get_text(strip=True)
            else:
                # just grab the 4th line of text if it exists
                location = lines[3] if len(lines) > 3 else ""

        return {
            'days': days,
            'times_12h': lines[2] if len(lines) > 2 else None,
            'location': location
        }
    
    # Parse Section Address
    try:
        parts = section_addr.split('.')
        match = re.match(r"([A-Za-z]+)(\d+)", parts[0])
        prefix = match.group(1)
        number = match.group(2)
        section = parts[1]
        term = parts[2]
    except Exception:
        print(f"Failed to parse section address: {section_addr}")
        prefix = number = section = term = ""

    # Enrollment
    curr = 0
    avail = 0
    wait = 0
    status_row_text = get_val("Status")
    try: 
        curr = int(re.search(r'Enrolled Total:\s*(-?\d+)', status_row_text).group(1)) if "Enrolled Total" in status_row_text else 0
        avail = int(re.search(r'Available Seats:\s*(-?\d+)', status_row_text).group(1)) if "Available Seats" in status_row_text else 0
        wait = int(re.search(r'Waitlist:\s*(-?\d+)', status_row_text).group(1)) if "Waitlist" in status_row_text else 0
    except Exception:
        print(f"Failed to parse enrollment info for {section_addr}: '{status_row_text}'")
    
    class_num_raw = ""
    people_data = {}
    schedule_data = {'days': [], 'times_12h': None, 'location': ""}
    school_name = None
    school_id = None
    try:
        people_data = parse_people()
        schedule_data = parse_location_and_schedule()
        class_num_raw = get_val("Class/Course Number")
        school_name, school_id = get_school()
    except Exception:
        print(f"Failed to parse")

    # Missing: topic, textbook?
    return {
        'section_address': section_addr,
        'course_prefix': prefix,
        'course_number': number,
        'section': section,
        'class_course_number': class_num_raw if class_num_raw else None,
        'class_level': get_val("Class Level"),
        'instruction_mode': get_val("Instruction Mode"),
        'title': soup.find('td', class_='courseinfo__overviewtable__coursetitle').get_text(strip=True),
        'description': get_val("Description"),
        'enrolled_status': get_enrolled_status(status_row_text),
        'enrolled_current': curr,
        'enrolled_max': curr + avail,
        'waitlist': wait,
        'term': term,
        'days': schedule_data['days'],
        'times_12h': schedule_data['times_12h'],
        'location': schedule_data['location'],
        'activity_type': get_val("Activity Type"),
        'semester_credit_hours': get_val("Semester Credit Hours"),
        'core': get_val("Core"),
        'grading': get_val("Grading"),
        'session_type': get_val("Session Type"),
        'add_consent': get_val("Add Consent"),
        'enrollment_reqs': get_list_val("Enrollment Reqs"),
        'class_attributes': get_list_val("Class Attributes"),
        'class_notes': get_val("Class Notes"),
        'instructors': people_data['instructors'],
        'instructor_ids': people_data['instructor_ids'],
        'tas': people_data['tas'],
        'ta_ids': people_data['ta_ids'],
        'school': school_name,
        'school_id': school_id,
        'syllabus': get_syllabus()
    }


# we have to click the overview button on each class to get waitlist cause report monkey doesn't give that info
def get_class_overview(data, session_id):
    data_json = json.loads(data)
    html_content = data_json["sethtml"]["#sr"]
    soup = BeautifulSoup(html_content, 'html.parser')

    rows = soup.find_all('tr', class_='cb-row')

    all_courses = []
    print(f"Getting overview for {len(rows)} classes")
    for i, row in enumerate(rows):
        section_address = row.get("data-id")
        data_req = row.get("data-req") # needed in request for overview
        row_id = row.get("id")
        div_id = f"{row_id}childcontent"
        
        overview_html, new_session_id = make_request_with_retry(
            get_class_detail, 
            session_id, 
            section_address, 
            data_req, 
            div_id
        )

        session_id = new_session_id

        width = len(str(len(rows)))
        print(f"({i+1:0{width}}/{len(rows)}): {section_address}")
        # with open(f"{section_address}.html", "w", encoding="utf-8") as f:
        #     f.write(overview_html)

        class_overview = parse_class_overview(overview_html, section_address)
        # print(f"{class_overview}")
        all_courses.append(class_overview)

    return all_courses


def get_text_or_none(out):
    if not out:
        return ""
    return out[0].text.strip()


def make_request_with_retry(request_func, session_id, *args, **kwargs):
    """
    Wraps a request function and retries on failure, refreshing the session ID.
    """
    max_retries = 3
    retries = 0
    current_session_id = session_id

    while retries < max_retries:
        try:
            response = request_func(current_session_id, *args, **kwargs)
            return response, current_session_id
        except (requests.exceptions.RequestException, Exception) as e:
            print(
                f'An error occurred: {e}. Retrying with a new session token...')
            retries += 1
            current_session_id = get_cookie()
            print(f'Attempt {retries}/{max_retries} with new session ID.')

    raise Exception(f'Failed to complete request after {max_retries} retries.')


def save_json(data, filters, option_value, term):
    filter_path = os.path.join(term, *filters.values()) if filters else term
    os.makedirs(filter_path, exist_ok=True)
    with open(os.path.join(filter_path, f"{option_value}.json"), "w", encoding="utf-8") as f:
        json.dump(data, f, indent=4)


def process_filters(session_id, term, all_data, dropdown_options, filters, filter_order, resume=None):
    """
    Recursively processes filters to scrape course data.
    """

    # base case: no more filters to apply
    if not filter_order:
        pass

    else:
        # get the next filter type and its options
        current_filter_type = filter_order[0]
        remaining_filter_order = filter_order[1:]

        options_key = FILTER_TYPES_MAP.get(current_filter_type)
        options = dropdown_options.get(options_key, [])

        # Optional: resume from a specific filter value
        if resume and not filters:
            print(f"Resuming from filter: {resume}")
            resume_index = options.index(resume)
            process_options = options[resume_index:]
        else:
            resume_index = 0
            process_options = options

        for i, option_value in enumerate(process_options):
            new_filters = filters.copy()
            new_filters[current_filter_type] = option_value
            print(
                f"[{resume_index+i+1}/{len(options)}] Processing {current_filter_type}: {option_value}")

            while True:
                try:
                    print(f"Making request with filters: {new_filters}")
                    response, session_id = make_request_with_retry(
                        make_course_request,
                        session_id,
                        term,
                        prefix=new_filters.get('prefix'),
                        school=new_filters.get('school'),
                        day=new_filters.get('day'),
                        level=new_filters.get('level')
                    )

                    if response.status_code != 200:
                        raise Exception('Failed to get the data page')

                    # if no items are found, continue to the next option
                    if '(no items found)' in response.text:
                        print('\tNo items found.')
                        save_json([], filters, option_value, term)
                        break

                    # if the query is too large, we break it down with more filters recursively by moving down the filter order
                    if 'displaying maximum' in response.text:
                        print(
                            f'\tQuery for {new_filters} results in more than 300 sections, splitting...')
                        session_id = process_filters(
                            session_id, term, all_data, dropdown_options, new_filters, remaining_filter_order)
                        break

                    class_overview = get_class_overview(response.text, session_id)

                    if class_overview:
                        save_json(class_overview, filters, option_value, term)

                    break

                except Exception as e:
                    print(f'Failed to get data for filters {new_filters}: {e}')
                    print('Attempting to get a new session token...')
                    session_id = get_cookie()
        return session_id


def combine_data(all_data, schools, prefixes, term):
    # Get first level directory names and .json files
    present = set()
    if os.path.exists(term):
        for entry in os.listdir(term):
            full_path = os.path.join(term, entry)
            if os.path.isdir(full_path):
                present.add(entry)
            elif entry.endswith(".json"):
                present.add(os.path.splitext(entry)[0])

    missing_schools = [s for s in schools if s not in present]
    missing_prefixes = [p for p in prefixes if p not in present]

    if missing_schools or missing_prefixes:
        print(f"Missing data for prefixes: {missing_prefixes})")
        print(f"Missing data for schools: {missing_schools}")
        return missing_prefixes, missing_schools

    for root, dirs, files in os.walk(term):
        for file in files:
            if file.endswith(".json"):
                with open(os.path.join(root, file), "r", encoding="utf-8") as f:
                    data = json.load(f)
                for d in data:
                    all_data[d['section_address']] = d

    return missing_prefixes, missing_schools
    

def scrape(session_id, term, resume):
    all_data = {}
    dropdown_ids = ['combobox_cp', 'combobox_col',
                    'combobox_days', 'combobox_clevel']
    dropdown_options = get_dropdown_options(dropdown_ids)

    dropdown_ids = [DROPDOWN_PREFIX_ID, DROPDOWN_SCHOOL_ID,
                    DROPDOWN_DAYS_ID, DROPDOWN_LEVELS_ID]
    dropdown_options = get_dropdown_options(dropdown_ids)

    if term == 'latest':
        term = get_latest_term()
        print(f'Using latest term: {term}')

    prefixes = dropdown_options.get(DROPDOWN_PREFIX_ID, [])
    schools = dropdown_options.get(DROPDOWN_SCHOOL_ID, [])
    days = dropdown_options.get(DROPDOWN_DAYS_ID, [])
    levels = dropdown_options.get(DROPDOWN_LEVELS_ID, [])

    if not prefixes or not schools:
        print("Could not retrieve all necessary dropdowns. Exiting.")
        return

    print(f'Found {len(prefixes)} prefixes, {len(schools)} schools, {len(days)} days, and {len(levels)} levels')


    if resume == "combine":
        print("RESUME set to 'combine'. Skipping scraping...")
    else:
        if resume not in prefixes and resume not in schools:
            print(f"RESUME '{resume}' not found in prefixes or schools. Starting from beginning.")
            resume = None
        elif resume:
            print(f"RESUME set to '{resume}'. Resuming from '{resume}'.")

        if resume not in schools:
            print("Processing prefixes...")
            session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                        ['prefix', 'day', 'level'],
                                        resume if resume in prefixes else None)
        else:
            print("Skipping prefix processing...")

        print("Processing schools...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['school', 'day', 'level'],
                                    resume if resume in schools else None)

    print(f"Combining data for {term}...")
    missing_prefixes, missing_schools = combine_data(all_data, schools, prefixes, term)
    if missing_prefixes or missing_schools:
        print("Missing filters. Exiting.")
        return

    final_data = list(all_data.values())
    print(f'\tGot {len(final_data)} unique classes for term {term}')

    out_dir = 'out'
    if not os.path.exists(out_dir):
        os.makedirs(out_dir)

    with open(f'{out_dir}/classes_{term}.json', 'w') as f:
        json.dump(final_data, f, indent=4)
        print(f"Data saved to {out_dir}/classes_{term}.json")
