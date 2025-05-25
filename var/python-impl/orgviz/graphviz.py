"""
The interface to graphviz, which at the moment, just means running it on the 
command line;
"""

import logging
import subprocess

def runDot(graphvizFilename, outputImageFilename, imageType="png"):
    try: 
        cmd = [
            "dot",
            "-T",
            imageType,
            graphvizFilename,
            "-o",
            outputImageFilename
        ]

        logging.debug("Running dot like this: %s", cmd)

        output = subprocess.run(cmd, shell=False, capture_output=True, check=False)

        stderr = output.stderr.decode('utf-8')
        stdout = output.stdout.decode('utf-8')

        logging.info("stderr: %s", stderr)
        logging.info("stdout: %s", stdout)

        if output.returncode == 0:
            logging.info("Completed sucessfully, rendered: %s", outputImageFilename)

            return True, None

        logging.error("dot output: %s", str(output))
    except FileNotFoundError: 
        logging.error("FileNotFoundError. Is the GraphViz's `dot` command installed on your computer? ")

        return False, "File not found"

    return False, stderr

