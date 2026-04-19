import html

import requests
import re
import json
import os
import time
from datetime import datetime
from bs4 import BeautifulSoup
from login import get_cookie
from parse import parse_class_overview

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

MAX_TIMEOUT = 1800  # 30 minutes
INITIAL_TIMEOUT = 60


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


# Get extra class overview detail
def make_overview_request(session_id, section_address, data_req, div_id):
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


def save_html(html, section_address, filters, term):
    filter_path = os.path.join(term, *filters.values())
    os.makedirs(filter_path, exist_ok=True)
    with open(os.path.join(filter_path, f"{section_address}.html"), "w", encoding="utf-8") as f:
        f.write(html)                                                                                               
                                                                                                                                                                                                                            

# we have to click the overview button on each class to get waitlist cause report monkey doesn't give that info
def get_class_overviews(data, session_id, filters, term):
    data_json = json.loads(data)
    html_content = data_json["sethtml"]["#sr"]
    soup = BeautifulSoup(html_content, 'html.parser')

    rows = soup.find_all('tr', class_='cb-row')

    failed_sections = []
    print(f"Getting overview for {len(rows)} classes")
    for i, row in enumerate(rows):
        section_address = row.get("data-id")
        data_req = row.get("data-req") # needed in request for overview
        row_id = row.get("id")
        div_id = f"{row_id}childcontent"

        overview_html, new_session_id = make_request_with_retry(
            make_overview_request,
            session_id,
            section_address,
            data_req,
            div_id
        )

        session_id = new_session_id

        width = len(str(len(rows)))
        print(f"({i+1:0{width}}/{len(rows)}): {section_address}")

        if not overview_html:
            print(f"Failed to get overview for {section_address}")
            failed_sections.append(section_address)
            continue

        save_html(overview_html, section_address, filters, term)

    return failed_sections


def get_text_or_none(out):
    if not out:
        return ""
    return out[0].text.strip()


def timeout_sleep(timeout):
    """
    Sleep for timeout seconds, double the next timeout (capped at MAX_TIMEOUT).
    """
    start_time = datetime.now().strftime('%H:%M:%S')
    print(f'Waiting {timeout}s before retry (started at {start_time})...')
    time.sleep(timeout)
    return min(timeout * 2, MAX_TIMEOUT)


def get_cookie_with_timeout(timeout=INITIAL_TIMEOUT):
    """
    Keep trying to get a session cookie with increasing timeout.
    """
    while True:
        try:
            return get_cookie(), timeout
        except Exception as e:
            print(f'Failed to get a new session token: {e}')
            timeout = timeout_sleep(timeout)


def make_request_with_retry(request_func, session_id, *args, **kwargs):
    """
    Wraps a request function and retries on failure, refreshing the session ID.
    """
    max_retries = 3
    current_session_id = session_id

    for attempt in range(1, max_retries + 1):
        try:
            response = request_func(current_session_id, *args, **kwargs)
            return response, current_session_id
        except Exception as e:
            print(f'An error occurred: {e}. Retrying with a new session token...')
            current_session_id = get_cookie()
            print(f'Attempt {attempt}/{max_retries} with new session ID.')

    raise Exception(f'Failed to complete request after {max_retries} retries.')


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

            timeout = INITIAL_TIMEOUT
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

                        # add dummy directory to show its scraped but empty
                        filter_path = os.path.join(term, *new_filters.values())
                        os.makedirs(filter_path, exist_ok=True)
                        break

                    # if the query is too large, we break it down with more filters recursively by moving down the filter order
                    if 'displaying maximum' in response.text:
                        print(
                            f'\tQuery for {new_filters} results in more than 300 sections, splitting...')
                        session_id = process_filters(
                            session_id, term, all_data, dropdown_options, new_filters, remaining_filter_order)
                        break

                    failed_courses = get_class_overviews(response.text, session_id, new_filters, term)

                    timeout = INITIAL_TIMEOUT
                    break

                except Exception as e:
                    print(f'Failed to get data for filters: {new_filters} (term: {term}): {e}')
                    timeout = timeout_sleep(timeout)
                    print('Attempting to get a new session token...')
                    session_id, timeout = get_cookie_with_timeout(timeout)

        return session_id


def combine_html(term, prefixes, schools):
    """Walk the term directory, parse every .html file, and return deduplicated course data."""
    if not os.path.exists(term):
        print(f"Directory '{term}' does not exist.")
        return None

    # Check if all prefix and school filters were scraped before combining
    present = set()
    for entry in os.listdir(term):
        if os.path.isdir(os.path.join(term, entry)):
            present.add(entry)

    missing_prefixes = [p for p in prefixes if p not in present]
    missing_schools = [s for s in schools if s not in present]

    if missing_prefixes or missing_schools:
        print(f"Missing data for prefixes: {missing_prefixes}")
        print(f"Missing data for schools: {missing_schools}")
        return None

    all_data = {}
    for root, dirs, files in os.walk(term):
        for file in files:
            if file.endswith(".html"):
                filepath = os.path.join(root, file)
                section_addr = os.path.splitext(file)[0]

                with open(filepath, "r", encoding="utf-8") as f:
                    html_content = f.read()

                section = parse_class_overview(html_content, section_addr)
                all_data[section['section_address']] = section

    return all_data
    

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

    elif resume in prefixes:
        print(f"Resuming from prefix '{resume}'.")
        print("Processing prefixes...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['prefix', 'day', 'level'], resume)
        print("Processing schools...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['school', 'day', 'level'])

    elif resume in schools:
        print(f"Resuming from school '{resume}'. Skipping prefix processing...")
        print("Processing schools...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['school', 'day', 'level'], resume)

    else:
        if resume:
            print(f"RESUME '{resume}' not found in prefixes or schools. Starting from beginning.")
        print("Processing prefixes...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['prefix', 'day', 'level'])
        print("Processing schools...")
        session_id = process_filters(session_id, term, all_data, dropdown_options, {},
                                    ['school', 'day', 'level'])

    print(f"Combining HTML data for {term}...")
    all_data = combine_html(term, prefixes, schools)
    if all_data is None:
        print("Missing filters. Exiting.")
        return

    final_data = list(all_data.values())
    print(f'\tGot {len(final_data)} unique classes for term {term}')

    out_dir = 'out'
    os.makedirs(out_dir, exist_ok=True)

    with open(f'{out_dir}/classes_{term}.json', 'w') as f:
        json.dump(final_data, f, indent=4)
        print(f"Data saved to {out_dir}/classes_{term}.json")
