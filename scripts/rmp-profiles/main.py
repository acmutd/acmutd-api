import json
import argparse
import time
import os
from scraper import scrape_rmp_data

def main():

    print("Scraping professor data from RateMyProfessors...")
    rmp_data = scrape_rmp_data(university_id="1273")

    print(f"Scraped data for {len(rmp_data)} professors.")

    out_dir = 'out'
    if not os.path.exists(out_dir):
        os.makedirs(out_dir)

    with open(f'{out_dir}/rmp_data.json', 'w') as f:
        json.dump(rmp_data, f, indent=4)
        print(f"Data saved to {out_dir}/rmp_data.json")


if __name__ == "__main__":
    main()