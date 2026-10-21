"""Waits for the org chart to render people before capture."""
import time

from selenium.webdriver.common.by import By


def run(driver):
    deadline = time.time() + 15
    while time.time() < deadline:
        if len(driver.find_elements(By.CSS_SELECTOR, "div.person")) >= 3:
            break
        time.sleep(0.25)
    names = [el.text for el in driver.find_elements(By.CSS_SELECTOR, "div.person")]
    if any("Bob Johnson" in name for name in names):
        # Node buttons are Alice, Bob, Charles, Imogen. Expand Bob and Imogen
        # so two reporting levels show; leave Charles collapsed so the cards
        # stay readable after fit().
        driver.execute_script(
            """
            const buttons = [...document.querySelectorAll('.node-button-g')];
            for (const i of [1, 3]) {
              if (buttons[i]) {
                buttons[i].dispatchEvent(new MouseEvent('click', {bubbles: true, view: window}));
              }
            }
            """
        )
        time.sleep(1.2)
    driver.execute_script(
        "if (window.chart && typeof window.chart.fit === 'function') { window.chart.fit(); }"
    )
    time.sleep(0.8)
    driver.execute_script("window.scrollTo(0, 0);")
