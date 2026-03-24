import re
from bs4 import BeautifulSoup


def clean(text):
    """Some classes use ∅ (U+2205) for empty fields."""
    return None if text == '\u2205' else text


def get_val(soup, label):
    """Finds label in header cells and returns the next td's text."""
    for th in soup.find_all('th'):
        if re.search(re.escape(label), th.get_text(), re.I):
            val_cell = th.find_next('td')
            return clean(val_cell.get_text(strip=True)) if val_cell else None
    for td in soup.find_all('td', class_='courseinfo__classsubtable__th'):
        if re.search(re.escape(label), td.get_text(), re.I):
            val_cell = td.find_next_sibling('td')
            return clean(val_cell.get_text(strip=True)) if val_cell else None
    return None


def get_list_val(soup, label):
    """Finds label in header cells and returns a list of <li> text values from the next td."""
    for th in soup.find_all('th'):
        if re.search(re.escape(label), th.get_text(), re.I):
            val_cell = th.find_next('td')
            if val_cell:
                return [li.get_text(strip=True) for li in val_cell.find_all('li')]
    return []


def parse_section_address(section_addr):
    """Returns (prefix, number, section, term) from a section address string."""
    try:
        parts = section_addr.split('.')
        course_id = parts[0]
        if re.search(r'\d', course_id):
            # Normal case: split at first digit (e.g. acct2301, mech6v49)
            match = re.match(r"([A-Za-z]+)(\d.*)", course_id)
            prefix = match.group(1)
            number = match.group(2)
        else:
            # all letter identifier: utd prefix (e.g. utdexcm)
            prefix = 'utd'
            number = course_id[3:]
        section = parts[1]
        term = parts[2]
        return prefix, number, section, term
    except Exception:
        print(f"Failed to parse section address: {section_addr}")
        return "", "", "", ""


def parse_enrollment(soup, section_addr):
    """Returns (status_text, enrolled_status, current, max, waitlist)."""
    status_text = get_val(soup, "Status")
    curr = 0
    avail = 0
    wait = 0
    try:
        curr = int(re.search(r'Enrolled Total:\s*(-?\d+)', status_text).group(1)) if status_text and "Enrolled Total" in status_text else 0
        avail = int(re.search(r'Available Seats:\s*(-?\d+)', status_text).group(1)) if status_text and "Available Seats" in status_text else 0
        wait = int(re.search(r'Waitlist:\s*(-?\d+)', status_text).group(1)) if status_text and "Waitlist" in status_text else 0
    except Exception:
        print(f"Failed to parse enrollment info for {section_addr}: '{status_text}'")

    if not status_text:
        enrolled_status = 'CLOSED'
    elif 'CANCELLED' in status_text:
        enrolled_status = 'CANCELLED'
    elif 'OPEN' in status_text:
        enrolled_status = 'OPEN'
    else:
        enrolled_status = 'CLOSED'

    return status_text, enrolled_status, curr, curr + avail, wait


def parse_people(soup):
    """Parse Instructors and TAs."""
    people = {
        'instructors': [], 'instructor_ids': [],
        'tas': [], 'ta_ids': []
    }

    for div in soup.select("div[id^='inst-']"):
        text_content = div.get_text(separator="・", strip=True)
        name = text_content.split("・")[0].strip()

        email_link = div.find("a", href=re.compile("mailto:"))
        net_id = email_link['href'].replace('mailto:', '').split('@')[0] if email_link else ""

        if "Teaching Assistant" in text_content:
            people['tas'].append(name)
            if net_id: people['ta_ids'].append(net_id)
        else:
            people['instructors'].append(name)
            if net_id: people['instructor_ids'].append(net_id)

    return people


def parse_schedule(soup, status_text, section_addr):
    """Gets location code and schedule details."""
    meeting_div = soup.find('div', class_='courseinfo__meeting-item--multiple')
    if not meeting_div:
        if status_text and 'CANCELLED' in status_text:
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


def parse_school(soup):
    """Returns (school_name, school_id)."""
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


def parse_syllabus(soup):
    """Returns the syllabus ID if found."""
    for th in soup.find_all('th'):
        if re.search(r'Syllabus', th.get_text(), re.I):
            td = th.find_next('td')
            if td:
                a = td.find('a', href=True)
                if a:
                    match = re.search(r'/(syl\w+)', a['href'])
                    return match.group(1) if match else None
    return None


def parse_class_overview(html, section_addr):
    """Parse a class detail HTML page into a course data dict."""
    soup = BeautifulSoup(html, "html.parser")

    prefix, number, section, term = parse_section_address(section_addr)
    status_text, enrolled_status, curr, enrolled_max, wait = parse_enrollment(soup, section_addr)
    people = parse_people(soup)
    schedule = parse_schedule(soup, status_text, section_addr)
    school_name, school_id = parse_school(soup)
    class_num_raw = get_val(soup, "Class/Course Number")

    return {
        'section_address': section_addr,
        'course_prefix': prefix,
        'course_number': number,
        'section': section,
        'class_course_number': class_num_raw if class_num_raw else None,
        'class_level': get_val(soup, "Class Level"),
        'instruction_mode': get_val(soup, "Instruction Mode"),
        'title': soup.find('td', class_='courseinfo__overviewtable__coursetitle').get_text(strip=True),
        'description': get_val(soup, "Description"),
        'enrolled_status': enrolled_status,
        'enrolled_current': curr,
        'enrolled_max': enrolled_max,
        'waitlist': wait,
        'term': term,
        'days': schedule['days'],
        'times_12h': schedule['times_12h'],
        'location': schedule['location'],
        'activity_type': get_val(soup, "Activity Type"),
        'semester_credit_hours': get_val(soup, "Semester Credit Hours"),
        'core': get_val(soup, "Core"),
        'grading': get_val(soup, "Grading"),
        'session_type': get_val(soup, "Session Type"),
        'add_consent': get_val(soup, "Add Consent"),
        'enrollment_reqs': get_list_val(soup, "Enrollment Reqs"),
        'class_attributes': get_list_val(soup, "Class Attributes"),
        'class_notes': get_val(soup, "Class Notes"),
        'instructors': people['instructors'],
        'instructor_ids': people['instructor_ids'],
        'tas': people['tas'],
        'ta_ids': people['ta_ids'],
        'school': school_name,
        'school_id': school_id,
        'syllabus': parse_syllabus(soup)
    }
