from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from os import environ
import time

url = 'https://coursebook.utdallas.edu'

def get_cookie():
    # Set up Selenium to use Chrome
    driver = None
    try:
        chrome_options = Options()
        chrome_options.add_argument("--log-level=3")
        chrome_options.add_argument("--ignore-certificate-errors")
        chrome_options.add_argument("--headless")
        # chrome_options.add_experimental_option("detach", True)  # Keeps the browser window open
        # service = Service("./chromedriver")  # Specify the path to your chromedriver
        driver = webdriver.Chrome(options=chrome_options)

    except Exception as e:
        print(f"Failed to start the Chrome driver: {e}")
        print("Go to https://googlechromelabs.github.io/chrome-for-testing/#stable to download the latest version of ChromeDriver. Copy the executable to the root folder of this project. You may also need the latest version of Chrome; make sure your chrome is updated.")
        exit(1)

    # Open a website of your choice
    driver.get(url)

    # Click the button with id 'pauth_link'
    try:
        # Wait for up to 10 seconds for the button to be clickable
        wait = WebDriverWait(driver, 10)
        button = wait.until(EC.element_to_be_clickable((By.ID, 'pauth_link')))
        
        # Click the button once it's clickable
        button.click()
        print("Button with ID 'pauth_link' clicked.")
    except Exception as e:
        print(f"Failed to find or click the button: {e}")
        exit(1)

    # Log in user
    try:
        # Wait for up to 10 seconds for the login form to be visible
        wait = WebDriverWait(driver, 10)
        netid_input = wait.until(EC.visibility_of_element_located((By.ID, 'netid')))
        password_input = wait.until(EC.visibility_of_element_located((By.ID, 'password')))
        print("Entering credentials...")
        netid_input.send_keys(environ['NETID'])
        password_input.send_keys(environ['PASSWORD'])
        pass
    except Exception as e:
        print(f"Failed to find the login form: {e}")
        exit(1)

    # Click the login button
    try:
        login_button = wait.until(EC.element_to_be_clickable((By.ID, 'login-button')))
        login_button.click()
        print("Login button clicked.")
    except Exception as e:
        print(f"Failed to click the login button: {e}")
        exit(1)

    # Wait until the login is complete
    try:
        # Wait for up to 60 seconds for the login complete to be visible
        wait = WebDriverWait(driver, 60)
        wait.until(EC.visibility_of_element_located((By.ID, 'guidedsearch')))
        print("Logged in successfully.")
    except Exception as e:
        print(f"Failed to login after 60 seconds: {e}")

    # Set the value of select element with id 'combobox_cp' to cp_acct
    driver.find_element(By.ID, 'combobox_cp').send_keys('cp_acct')

    # Click the button with the text "Search Classes"
    driver.find_element(By.XPATH, '//button[text()="Search Classes"]').click()

    # Reload window
    driver.refresh()

    # Wait 1 second
    time.sleep(1)

    # Retrieve cookies from the browser session
    cookies = driver.get_cookies()

    # Get the PTGSESSID from the cookies
    session_id = None
    for cookie in cookies:
        if cookie['name'] == 'PTGSESSID':
            session_id = cookie['value']
            break
    if session_id is None:
        print('Failed to find the PTGSESSID (login token) cookie')
        exit(1)

    # Quit the browser after retrieving cookies (optional)
    driver.quit()

    print(f"Session ID: {session_id}")

    return session_id
