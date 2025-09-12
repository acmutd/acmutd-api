from login import get_cookie
from grab_data import scrape
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

    # Get the terms (comma-separated)
    terms = [term.strip() for term in environ['CLASS_TERMS'].split(',') if term.strip()]

    # Call the function to get the cookie
    session_id = get_cookie()

    # Loop over each term and scrape data
    for term in terms:
        print(f"Scraping data for term: {term}")
        scrape(session_id, term)

if __name__ == '__main__':
    main()
