from login import get_cookie
from grab_data import get_prefixes, scrape
from os import environ
import dotenv

# Load .env file
dotenv.load_dotenv()

def main():
    # Check for environmental variables
    if 'CLASS_TERM' not in environ:
        print("CLASS_TERM environmental variable not set.")
        exit(1)
    if 'NETID' not in environ:
        print("NETID environmental variable not set.")
        exit(1)
    if 'PASSWORD' not in environ:
        print("PASSWORD environmental variable not set.")
        exit(1)

    # Get the term
    term = environ['CLASS_TERM']

    # Call the function to get the cookie
    session_id = get_cookie()

    # GET ALL THE DATA!!
    scrape(session_id, term)

if __name__ == '__main__':
    main()
