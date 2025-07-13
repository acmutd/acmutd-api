from login import get_cookie
from grab_data import get_prefixes, get_terms, scrape
from save_data import save_data
from os import environ
import dotenv

# Load .env file
dotenv.load_dotenv()


def main():
    # Check for environmental variables
    if 'CLASS_TERMS' not in environ:
        print("CLASS_TERMS environmental variable not set.")
        exit(1)
    if 'NETID' not in environ:
        print("NETID environmental variable not set.")
        exit(1)
    if 'PASSWORD' not in environ:
        print("PASSWORD environmental variable not set.")
        exit(1)

    # Get prefixes
    prefixes = get_prefixes()

    # Get all available terms and user defined term list
    valid_terms = get_terms()
    terms = str_to_list(environ['CLASS_TERMS'])

    # Loop through each term
    for term in terms:
        # Skip invalid terms
        if term not in valid_terms:
            print(f"Term {term} is not valid. Skipping...")
            continue

        # Call the function to get the cookie
        session_id = get_cookie()

        # GET ALL THE DATA!!
        data = scrape(session_id, term, prefixes)

        # Write data to local file
        save_data(data, '/app/output', term)


def str_to_list(s):
    """
    Convert a string to a list of strings.
    """
    return list(map(lambda x: x.strip(), s.split(',')))


if __name__ == '__main__':
    main()
