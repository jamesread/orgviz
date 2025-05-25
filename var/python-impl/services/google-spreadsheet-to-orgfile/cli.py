#!/usr/bin/env python

import configargparse

from lib import setCredentialsJson, generateDot

parser = configargparse.ArgumentParser(default_config_files=["~/.spreadsheet-reader.cfg"]);
parser.add_argument("--credentialsJson", required=True)
parser.add_argument("--spreadsheetId", required=True)
args = parser.parse_args();

def main():
    setCredentialsJson(args.credentialsJson)
    out = generateDot(args.spreadsheetId)

    print (out)

if __name__ == '__main__':
    main()
