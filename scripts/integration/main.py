# integration scraper will read from /in to retrieve coursebook data, grades data, and rmp data
# then will pretty much follow original professor scraper logic:
# 1. run aggregator.py logic to aggregate grades data across every semester
# 2. run professor main.py logic to match rmp data to aggregated grades data (matched data will use normalized names as the key)
# 3. match coursebook data to original grades data to get instructor ids
# 4. for remaining grades data without coursebook sections, try to match through the instructor name on the professor data
# 5. copy instructor data to a new set with id as the key
# 6. return everything to /out files for go driver to upload to the desired save environment

def main():
   pass

if __name__ == "__main__":
    main()