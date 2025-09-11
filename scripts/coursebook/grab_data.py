import requests
import re
import json
from bs4 import BeautifulSoup
from login import get_cookie

base_url = 'https://coursebook.utdallas.edu'
url = 'https://coursebook.utdallas.edu/clips/clip-cb11-hat.zog'
output = 'classes.json'

def get_prefixes():
    res = requests.get(base_url)

    if res.status_code != 200:
        print('Failed to get coursebook website')
        print(res.text)
        exit(1)

    matches = re.findall(r'\<option value="cp_acct.*\<\/select\>', res.text)
    raw_pre = matches[0]
    
    # Use regex to extract all value fields
    values = re.findall(r'value="([^"]+)"', raw_pre)

    return values

def get_schools():
    res = requests.get(base_url)

    if res.status_code != 200:
        print('Failed to get coursebook website (schools)')
        print(res.text)
        exit(1)

    matches = re.findall(r'<select class="combobox search-phrase" id="combobox_col">.*?</select>', res.text, re.S)
    if not matches:
        print("Failed to find schools dropdown")
        return []

    raw_schools = matches[0]
    values = re.findall(r'value="([^"]+)"', raw_schools)
    values = [v for v in values if v.strip() and not v.lower().startswith("any")]

    return values

def get_days():
    res = requests.get(base_url)

    if res.status_code != 200:
        print('Failed to get coursebook website (days)')
        print(res.text)
        exit(1)

    matches = re.findall(r'<select class="combobox search-phrase" id="combobox_days">.*?</select>', res.text, re.S)
    if not matches:
        print("Failed to find days dropdown")
        return []

    raw_days = matches[0]
    values = re.findall(r'value="([^"]+)"', raw_days)
    values = [v for v in values if v.strip() and not v.lower().startswith("any")]

    return values


def make_course_request(session_id, term, prefix, day=None):
    """
    Perform a POST to coursebook with given filters and return the response.
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

    if day is None:
        data = {
            'action': 'search',
            's[]': [f'term_{term}', prefix]
        }
    else: 
        data = {
            'action': 'search',
            's[]': [f'term_{term}', prefix, day]
        }

    response = requests.post(url, headers=headers, data=data, timeout=5)
    if response.status_code != 200:
        raise Exception(f"Failed course request: {response.text[:200]}")

    return response

def make_monkey_request(report_id, session_id):
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

def scrape(session_id, term):
    # Keep track of all data
    all_data = []

    prefixes = get_prefixes()
    schools = get_schools()
    days = get_days()

    print(f'Found {len(prefixes)} prefixes and {len(schools)} schools and {len(days)} days')

    seen_sections = set() # used to avoid duplicates when gathering from multiple filters

    # Loop through all the classes
    for i,p in enumerate(prefixes):
        while True:
            try:
                print(f'[{i+1}/{len(prefixes)}] Getting data for {term} prefix {p}')

                # Get the response
                response = make_course_request(session_id, term, p)

                if response.status_code != 200:
                    print('Failed to get the data page')
                    print(response)
                    print(response.text)
                    raise Exception('Failed to get the data page')

                # If 0 items
                if '(no items found)' in response.text:
                    print('\tNo items found')
                    break

                # If the "displaying maximum" text is found, need to do individual requests
                # on the day modality to split up data
                if 'displaying maximum' in response.text:
                    print('\tCurrent term has more than 300 items, need to split up data query')
                    new_data = find_big_term_prefix(p, term, session_id, days)
                    for d in new_data:
                        if d['section_address'] not in seen_sections:
                            all_data.append(d)
                            seen_sections.add(d['section_address'])
                    break

                # Get number of items
                items = re.findall(r'(\d+)\s*item(?:s)?', response.text)
                if len(items) == 0:
                    print('\tFailed to find number of items')
                    raise Exception('Failed to find number of items')

                items = int(items[0])
                print(f"Number of classes detected in coursebook: {items}")

                if items == 0:
                    print('\tNo items found')
                    break
                elif items == 1:
                    class_data = get_single_class(response.text, term, p)
                    if class_data['section_address'] not in seen_sections:
                        all_data.append(class_data)
                        seen_sections.add(class_data['section_address'])
                    break

                # Use the regex to find the desired part of the response
                matches = re.findall(r'\/reportmonkey\\\/cb11-export\\\/(.*?)\\\"', response.text)

                # Print the matched results
                if len(matches) == 0:
                    print('Failed to find the report ID from the response:')
                    print(response.text)
                    raise Exception('Failed to find the report ID from the response')
                report_id = matches[-1]

                monkey_response = make_monkey_request(report_id, session_id)

                if monkey_response.status_code != 200:
                    print('Failed to get the report response')
                    print(monkey_response.text)
                    raise Exception('Failed to get the report response')

                new_data = monkey_response.json()

                # Get the instructor netids
                # and append the instructor ids to the data
                ids, names = get_instructor_netids(response.text)
                # Fallback: If report_data is missing or empty, manually extract each class
                report_data = new_data.get('report_data', [])
                if len(report_data) != items:
                    print(f'\tWarning: Number of classes in report data ({len(report_data)}) does not match expected ({items})')
                if not report_data:
                    print('\tReport monkey returned no classes, manually extracting each class...')
                    # Parse the HTML from response.text
                    try:
                        data_json = json.loads(response.text)
                        html_content = data_json["sethtml"]["#sr"]
                        soup = BeautifulSoup(html_content, 'html.parser')
                        # Each course row typically has a clickable link with class 'stopbubble'
                        course_links = soup.find_all('a', class_='stopbubble')
                        for link in course_links:
                            # For each course, get its HTML block and pass to get_single_class
                            # Find the parent row (tr) and get its HTML
                            parent_row = link.find_parent('tr')
                            if parent_row:
                                # Build a mini HTML table for this row
                                mini_html = str(parent_row)
                                # Wrap in the expected JSON structure for get_single_class
                                mini_json = json.dumps({"sethtml": {"#sr": mini_html}})
                                class_data = get_single_class(mini_json, term, p)
                                all_data.append(class_data)
                    except Exception as e:
                        print(f'Failed to manually extract classes: {e}')
                else:
                    for j, d in enumerate(report_data):
                        d['instructors'] = names[j] if j < len(names) else ''
                        d['instructor_ids'] = ids[j] if j < len(ids) else ''
                        if d['section_address'] not in seen_sections:
                            all_data.append(d)
                            seen_sections.add(d['section_address'])
                break
            except Exception as e:
                print(f'Failed to get data for prefix {p}: {e}')
                print(f'Prompting for new token...')
                session_id = get_cookie()


    print(f"\nChecking {len(schools)} schools for missing classes...")
    for s in schools:
        while True:
            try:
                print(f'Checking school {s}')
                response = make_course_request(session_id, term, s)

                if '(no items found)' in response.text:
                    break

                if 'displaying maximum' in response.text:
                    print(f'\tSchool {s} query exceeds 300 items, splitting by days...')
                    new_data = find_big_term_school(s, term, session_id, days)
                    for d in new_data:
                        if d['section_address'] not in seen_sections:
                            print(f"\tAdding missing class {d['section_address']} from {s}")
                            all_data.append(d)
                            seen_sections.add(d['section_address'])
                    break

                items = re.findall(r'(\d+)\s*item(?:s)?', response.text)
                items = int(items[0]) if items else 0
                if items == 0:
                    break

                if items == 1:
                    class_data = get_single_class(response.text, term, s)
                    if class_data['section_address'] not in seen_sections:
                        print(f"\tAdding missing class {class_data['section_address']} from {s}")
                        all_data.append(class_data)
                        seen_sections.add(class_data['section_address'])
                    break

                matches = re.findall(r'\/reportmonkey\\\/cb11-export\\\/(.*?)\\\"', response.text)
                if not matches:
                    break
                report_id = matches[-1]
                monkey_response = make_monkey_request(report_id, session_id)
                report_data = monkey_response.json().get('report_data', [])

                ids, names = get_instructor_netids(response.text)
                for j, d in enumerate(report_data):
                    d['instructors'] = names[j] if j < len(names) else ''
                    d['instructor_ids'] = ids[j] if j < len(ids) else ''
                    if d['section_address'] not in seen_sections:
                        print(f"\tAdding missing class {d['section_address']} from {s}")
                        all_data.append(d)
                        seen_sections.add(d['section_address'])
                break

            except Exception as e:
                print(f'Failed to get data for school {s}: {e}')
                session_id = get_cookie()


    # Write the data to a file
    print(f'\tGot {len(all_data)} classes for term {term}')
    with open(f'classes_{term}.json', 'w') as f:
        json.dump(all_data, f, indent=4)


# THIS IS FOR A STUPID EDGE CASE
# Coursebook has a limit of 300 items per query
# If the number of items is greater than 300, we need to split up the query
# into individual days
def find_big_term_prefix(prefix, term, session_id, days):
    all_data = []
    for i, day in enumerate(days):
        while True:
            try:
                print(f'\t[{i+1}/{len(days)}] Getting data for prefix {prefix} ({day})')

                # Get the response
                # response = requests.post(url, headers=headers, data=data, timeout=5)
                response = make_course_request(session_id, term, prefix, day)

                if response.status_code != 200:
                    print('Failed to get the data page')
                    print(response)
                    print(response.text)
                    raise Exception('Failed to get the data page')

                # If 0 items
                if '(no items found)' in response.text:
                    print('\tNo items found')
                    break

                # If the "displaying maximum" text is found, something is wrong...
                if 'displaying maximum' in response.text:
                    print('ERROR: Query with term, prefix, and day still exceeds 300 items. This should not happen.')
                    print(f'Term {term}, Prefix {prefix}, Day {day}')
                    exit(1)

                # Get number of items
                items = re.findall(r'(\d+)\s*item(?:s)?', response.text)
                if len(items) == 0:
                    print('\tFailed to find number of items')
                    raise Exception('Failed to find number of items')

                items = int(items[0])

                if items == 0:
                    print('\tNo items found')
                    break
                elif items == 1:
                    class_data = get_single_class(response.text, term, prefix)
                    all_data.append(class_data)
                    break

                # Use the regex to find the desired part of the response
                matches = re.findall(r'\/reportmonkey\\\/cb11-export\\\/(.*?)\\\"', response.text)

                # Print the matched results
                if len(matches) == 0:
                    print('Failed to find the report ID from the response:')
                    print(response.text)
                    raise Exception('Failed to find the report ID from the response')
                report_id = matches[-1]

                monkey_response = make_monkey_request(report_id, session_id)

                if monkey_response.status_code != 200:
                    print('Failed to get the report response')
                    print(monkey_response.text)
                    raise Exception('Failed to get the report response')

                new_data = monkey_response.json()

                # Get the instructor netids
                # and append the instructor ids to the data
                ids, names = get_instructor_netids(response.text)
                for i, d in enumerate(new_data['report_data']):
                    d['instructors'] = names[i]
                    d['instructor_ids'] = ids[i]
                
                all_data.extend(new_data['report_data'])
                break
            except Exception as e:
                print(f'Failed to get data for prefix {prefix}, day {day}: {e}')
                print(f'Prompting for new token...')
                session_id = get_cookie()
    return all_data

def find_big_term_school(school, term, session_id, days):
    all_data = []
    for i, day in enumerate(days):
        while True:
            try:
                print(f'\t[{i+1}/{len(days)}] Getting data for school {school} ({day})')

                response = make_course_request(session_id, term, school, day)

                if response.status_code != 200:
                    raise Exception('Failed to get the data page')

                if '(no items found)' in response.text:
                    print('\tNo items found')
                    break

                if 'displaying maximum' in response.text:
                    print('ERROR: Query with term, school, and day still exceeds 300 items. This should not happen.')
                    print(f'Term {term}, School {school}, Day {day}')
                    exit(1)

                items = re.findall(r'(\d+)\s*item(?:s)?', response.text)
                items = int(items[0]) if items else 0
                if items == 0:
                    break
                elif items == 1:
                    class_data = get_single_class(response.text, term, school)
                    all_data.append(class_data)
                    break

                matches = re.findall(r'\/reportmonkey\\\/cb11-export\\\/(.*?)\\\"', response.text)
                if not matches:
                    raise Exception('Failed to find report ID from the response')
                report_id = matches[-1]

                monkey_response = make_monkey_request(report_id, session_id)
                if monkey_response.status_code != 200:
                    raise Exception('Failed to get the report response')

                new_data = monkey_response.json().get('report_data', [])
                ids, names = get_instructor_netids(response.text)

                for j, d in enumerate(new_data):
                    d['instructors'] = names[j] if j < len(names) else ''
                    d['instructor_ids'] = ids[j] if j < len(ids) else ''
                    all_data.append(d)
                break
            except Exception as e:
                print(f'Failed to get data for school {school}, day {day}: {e}')
                print(f'Prompting for new token...')
                session_id = get_cookie()
    return all_data


# If only one class is found, no report monkey thing
# is generated, which means we have to MANUALLY find the data smh
def get_single_class(data, term, prefix):
    # Parse the string as JSON to get the HTML part
    data_json = json.loads(data)
    html_content = data_json["sethtml"]["#sr"]

    # Parse the HTML using BeautifulSoup
    soup = BeautifulSoup(html_content, 'html.parser')

    # Extract the required fields
    class_section = get_text_or_none(soup.find_all('a', class_='stopbubble'))
    class_title = get_text_or_none(soup.find_all('td', style="line-height: 1.1rem;")).strip()
    schedule_day = get_text_or_none(soup.find_all('span', class_='clstbl__resultrow__day'))
    schedule_time = get_text_or_none(soup.find_all('span', class_='clstbl__resultrow__time'))
    location = get_text_or_none(soup.find_all('div', class_='clstbl__resultrow__location'))

    # Split the section string up
    a = class_section.split(" ")
    b = a[1].split(".") if len(a) >= 2 else a[0].split(".")
    number = b[0]
    section = b[1]
    section_addr = class_section.replace(' ', '').lower() + '.' + term

    # EDGE EDGE CASE
    # For all "utd" prefixes, the course number is always "STAB" even though it doesn't show in the UI...
    if prefix == 'utd' and number == '':
        number = 'stab'
        section_addr = 'utdstab' + section_addr

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
        matches = re.findall(r'http:\/\/coursebook.utdallas.edu\/search\/(.*?)" title="(.*?)"', str(row))
        if len(matches) == 0:
            netids.append('')
            names.append('')
            continue
        match_zipped = list(zip(*matches))
        netids.append(', '.join(match_zipped[0]))
        names.append(', '.join(match_zipped[1]))
    
    return netids, names


def get_text_or_none(out):
    if not out:
        return ""
    return out[0].text.strip()
