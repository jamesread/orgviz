"""Switches the header to the Second Company chart."""
import time

from selenium.webdriver.common.by import By


def run(driver):
    for link in driver.find_elements(By.CSS_SELECTOR, "#charts-list a"):
        if "Second Company" in link.text:
            link.click()
            break
    deadline = time.time() + 10
    while time.time() < deadline:
        texts = [el.text for el in driver.find_elements(By.CSS_SELECTOR, "div.person")]
        if any("CFO" in text or "Founder" in text for text in texts):
            break
        time.sleep(0.25)
    driver.execute_script(
        "if (window.chart && typeof window.chart.fit === 'function') { window.chart.fit(); }"
    )
    time.sleep(0.5)
    driver.execute_script("window.scrollTo(0, 0);")
