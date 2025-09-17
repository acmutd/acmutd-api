need to do the following:
1. add folder for user to input the .xlsx file
2. create python handling for converting to csv based on the term defined in the env
    i.e. TERM=24s --> file name = "Spring 2024" (s = Spring, u = Summer, f = Fall, + `20{first part of term}`)
3. resulting data should be downloaded for now, uploaded to firebase
4. should build some kind of 4th "scraper" integration service that takes the available grades data from a given term and does the following:
    1. pulls the coursebook data from that term (scrape new if not saved already)
    2. add the instructor id to the grade distribution for each respective section found in coursebook
    3. 