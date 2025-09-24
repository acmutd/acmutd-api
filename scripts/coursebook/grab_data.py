import requests
import re
import json
from bs4 import BeautifulSoup
from login import get_cookie

base_url = 'https://coursebook.utdallas.edu'
url = 'https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog'
output = 'classes.json'

DROPDOWN_PREFIX_ID = 'combobox_cp'
DROPDOWN_SCHOOL_ID = 'combobox_col'
DROPDOWN_DAYS_ID = 'combobox_days'
DROPDOWN_LEVELS_ID = 'combobox_clevel'

FILTER_TYPES_MAP = {
    'prefix': DROPDOWN_PREFIX_ID,
    'school': DROPDOWN_SCHOOL_ID,
    'day': DROPDOWN_DAYS_ID,
    'level': DROPDOWN_LEVELS_ID,
}


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


def get_instructor_netids(data):
    # Parse the string as JSON to get the HTML part
    data_json = json.loads(data)
    html_content = data_json["sethtml"]["#sr"]

    # Parse the HTML using BeautifulSoup
    soup = BeautifulSoup(html_content, 'html.parser')

    # Extract the netid field
    rows = soup.find_all('tr', class_='cb-row')
    netids = []
    names = []
    for row in rows:
        matches = re.findall(
            r'http:\/\/coursebook.utdallas.edu\/search\/(.*?)" title="(.*?)"', str(row))
        if len(matches) == 0:
            netids.append('')
            names.append('')
            continue
        match_zipped = list(zip(*matches))
        netids.append(', '.join(match_zipped[0]))
        names.append(', '.join(match_zipped[1]))

    return netids, names


# If only one class is found, no report monkey thing
# is generated, which means we have to MANUALLY find the data smh
def get_single_class(data, term, filter):
    # Parse the string as JSON to get the HTML part
    data_json = json.loads(data)
    html_content = data_json["sethtml"]["#sr"]

    # Parse the HTML using BeautifulSoup
    soup = BeautifulSoup(html_content, 'html.parser')

    # Extract the required fields
    class_section = get_text_or_none(soup.find_all('a', class_='stopbubble'))
    class_title = get_text_or_none(soup.find_all(
        'td', style="line-height: 1.1rem;")).strip()
    schedule_day = get_text_or_none(soup.find_all(
        'span', class_='clstbl__resultrow__day'))
    schedule_time = get_text_or_none(soup.find_all(
        'span', class_='clstbl__resultrow__time'))
    location = get_text_or_none(soup.find_all(
        'div', class_='clstbl__resultrow__location'))

    # Parse prefix, number, section from class_section
    # Example: "ACCT 2301.001"
    prefix, number, section = '', '', ''
    try:
        parts = class_section.split()
        if len(parts) == 2:
            prefix = parts[0]
            num_sec = parts[1].split('.')
            if len(num_sec) == 2:
                number = num_sec[0]
                section = num_sec[1]
    except Exception:
        print(f"Failed to parse class section: {class_section}")

    section_addr = f"{prefix.lower()}{number}.{section}{term}"

    # EDGE EDGE CASE
    # For all "utd" prefixes, the course number is always "STAB" even though it doesn't show in the UI...
    if (filter == 'utd' or prefix == 'utd') and number == '':
        number = 'stab'
        # might be section_addr = 'utdstab' + section_addr
        section_addr = f"utdstab.{section}{term}"

    print(f"Parsed single class: {prefix} {number}.{section}")

    # Get the instructor netid
    instructor_netids, instructors = get_instructor_netids(data)
    if len(instructor_netids) == 0:
        instructors = ['']
        instructor_netids = ['']

    # Return the extracted values
    return {
        'section_address': section_addr,
        'course_prefix': prefix,
        'course_number': number,
        'section': section,
        'title': class_title.replace(r'\(.*\)', ''),
        'term': term,
        'instructors': instructors[0],
        'instructor_ids': instructor_netids[0],
        'days': schedule_day.replace(' & ', ','),
        'times_12h': schedule_time,
        'location': location
    }


def manually_parse_html_data(html_content, term, filter_value, all_data):
    """
    Manually extracts and processes class data from raw HTML content.
    """
    try:
        data_json = json.loads(html_content)
        html_content = data_json["sethtml"]["#sr"]
    except json.JSONDecodeError:
        pass

    soup = BeautifulSoup(html_content, 'html.parser')
    course_links = soup.find_all('a', class_='stopbubble')
    for link in course_links:
        parent_row = link.find_parent('tr')
        if parent_row:
            mini_html = str(parent_row)
            mini_json = json.dumps({"sethtml": {"#sr": mini_html}})
            class_data = get_single_class(mini_json, term, filter_value)
            all_data[class_data['section_address']] = class_data


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


def process_filters(session_id, term, all_data, dropdown_options, filters, filter_order):
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

        for i, option_value in enumerate(options):
            new_filters = filters.copy()
            new_filters[current_filter_type] = option_value
            print(
                f"[{i+1}/{len(options)}] Processing {current_filter_type}: {option_value}")

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
                        break

                    # if the query is too large, we break it down with more filters recursively by moving down the filter order
                    if 'displaying maximum' in response.text:
                        print(
                            f'\tQuery for {new_filters} results in more than 300 sections, splitting...')
                        session_id = process_filters(
                            session_id, term, all_data, dropdown_options, new_filters, remaining_filter_order)
                        break

                    # check if there is only one item (report monkey download link not generated) --> nvm i handled this with the if/else below
                    items = re.findall(r'(\d+)\s*item(?:s)?', response.text)
                    items = int(items[0]) if items else 0

                    # try to get the report monkey endpoint to get the JSON data
                    print(
                        f'\tFound {items} classes for filters {new_filters}. Attempting to get report...')
                    matches = re.findall(
                        r'\/reportmonkey\\\/cb11-export\\\/(.*?)\\\"', response.text)
                    report_data = []

                    if matches:
                        report_id = matches[-1]

                        try:
                            monkey_response, session_id = make_request_with_retry(
                                make_monkey_request, session_id, report_id)
                            new_data = monkey_response.json()
                            report_data = new_data.get('report_data', [])

                        except json.JSONDecodeError:
                            print(
                                f'\tFailed to get report monkey data for {new_filters}: Expecting value: line 1 column 1 (char 0). Falling back to HTML parsing.')

                        # if the report data is empty or missing then we scrape the html from coursebook, happens for certain reports of size 2-3 sometimes
                        if not report_data:
                            print(
                                '\tReport monkey returned no classes, manually extracting each class from original HTML...')
                            manually_parse_html_data(response.text, term, new_filters.get(
                                'prefix', new_filters.get('school')), all_data)
                        else:
                            # If the report data is valid, process it
                            print(
                                f'\tSuccessfully retrieved {len(report_data)} classes from report.')
                            ids, names = get_instructor_netids(response.text)
                            for j, d in enumerate(report_data):
                                d['instructors'] = names[j] if j < len(
                                    names) else ''
                                d['instructor_ids'] = ids[j] if j < len(
                                    ids) else ''
                                all_data[d['section_address']] = d
                    else:
                        # if the report is missing, we have to manually scrape the html from coursebook, happens when there is only 1 class found
                        print(
                            f'\tFailed to find report ID from the response. Manually extracting...')
                        manually_parse_html_data(response.text, term, new_filters.get(
                            'prefix', new_filters.get('school')), all_data)

                    # Break the while loop if a valid response was processed
                    break

                except Exception as e:
                    print(f'Failed to get data for filters {new_filters}: {e}')
                    print('Attempting to get a new session token...')
                    session_id = get_cookie()
        return session_id


def scrape(session_id, term):
    all_data = {}
    dropdown_ids = ['combobox_cp', 'combobox_col',
                    'combobox_days', 'combobox_clevel']
    dropdown_options = get_dropdown_options(dropdown_ids)

    dropdown_ids = [DROPDOWN_PREFIX_ID, DROPDOWN_SCHOOL_ID,
                    DROPDOWN_DAYS_ID, DROPDOWN_LEVELS_ID]
    dropdown_options = get_dropdown_options(dropdown_ids)

    prefixes = dropdown_options.get(DROPDOWN_PREFIX_ID, [])
    schools = dropdown_options.get(DROPDOWN_SCHOOL_ID, [])
    days = dropdown_options.get(DROPDOWN_DAYS_ID, [])
    levels = dropdown_options.get(DROPDOWN_LEVELS_ID, [])

    if not prefixes or not schools:
        print("Could not retrieve all necessary dropdowns. Exiting.")
        return

    print(f'Found {len(prefixes)} prefixes, {len(schools)} schools, {len(days)} days, and {len(levels)} levels')

    print("processing prefixes")
    session_id = process_filters(session_id, term, all_data, dropdown_options, {}, [
                                 'prefix', 'day', 'level'])
    print("processing schools")
    session_id = process_filters(session_id, term, all_data, dropdown_options, {}, [
                                 'school', 'day', 'level'])

    final_data = list(all_data.values())
    print(f'\tGot {len(final_data)} unique classes for term {term}')

    import os
    out_dir = 'out'
    if not os.path.exists(out_dir):
        os.makedirs(out_dir)

    with open(f'{out_dir}/classes_{term}.json', 'w') as f:
        json.dump(final_data, f, indent=4)
        print(f"Data saved to {out_dir}/classes_{term}.json")
