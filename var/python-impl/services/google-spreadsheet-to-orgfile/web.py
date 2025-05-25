#!/usr/bin/env python3

import cherrypy

import configargparse
import os
import json

from time import sleep

import logging
import sys
from sys import exc_info
from googleapiclient.errors import HttpError

from lib import generateDot, getSheetsApi, setCredentialsJson, setCookieFilename

parser = configargparse.ArgumentParser(default_config_files = ["~/.spreadsheet-reader-web.cfg"])
parser.add_argument("--logging", type = int, default = 20, help = "1 = Everything. 50 = Critical only.", env_var = "LOGGING")
parser.add_argument("--credentialsJson", default = '/etc/orgviz/spreadsheet-reader-credentials.json', env_var = 'CREDENTIALS_JSON')
parser.add_argument("--cookieFilename", default = '/etc/orgviz-cookie/cookie', env_var = 'COOKIE_FILENAME')
parser.add_argument('--port', default = 8081, type = int, env_var = "PORT");
args = parser.parse_args();

def jsonHttpErrorResponse(errorMessage):
    cherrypy.response.headers['Content-Type'] = 'application/json'
    cherrypy.response.status = 400;

    logging.error(errorMessage)

    return json.dumps({
        "errorMessage": errorMessage
    }, ensure_ascii=False).encode('utf8')

class FrontendWrapper:
    @cherrypy.expose
    def index(self):
        ret = "";
        ret += '<h1>orgviz Spreadsheet Reader Web Service</h1>'
        ret += '<li>GET <a href = "/generateFromSheet/spreadsheetId"><tt>/generateFromSheet/<strong>yourSpreadsheetId</strong></tt></a></li>'


        return ret

    def generateDotAndCatchError(self, spreadsheetId): 
        ret = ""
        errorMessage = None

        try:
            ret = generateDot(spreadsheetId)
        except HttpError as e:
            if e.resp.status == 404:
                errorMessage = "Google spreadsheet not found with that ID"
            elif e.resp.status == 403:
                errorMessage = "Cannot access that Google spreadsheet, check the sharing settings."
            else:
                errorMessage = "Unkown HTTP error: " + str(e.resp.status)
        except Exception as e:
            errorMessage = str(type(e)) + ": " +  str(e)

        return ret, errorMessage

    @cherrypy.expose
    def generateFromSheet(self, spreadsheetId="none"):
        ret, errorMessage = self.generateDotAndCatchError(spreadsheetId)

        if errorMessage is not None:
            logging.error("generateFromSheet: " + errorMessage)

            return jsonHttpErrorResponse(errorMessage)
        else:
            cherrypy.response.headers['Content-Type'] = 'text/plain'

            return ret

def setupCredentialsJson():
    # Design choice: sleep-wait for the file, as containers might take a short
    # time to "mount" the configuration file, or it might not be configured, 
    # and without it the app will just crash/exit later. Hence, sleep-wait 
    # instead if hard-fail-fast.

    while not os.path.exists(args.credentialsJson):
        logging.info("Credentials file does not exist, expected it here: " + args.credentialsJson)
        sleep(10)

    logging.info("Found credentials file: " + args.credentialsJson)

    setCredentialsJson(args.credentialsJson)

    # Behavior: Get the sheets API here (and do nothing with it) so that the 
    # first "real" request does not stall while the login flow is completed.

    getSheetsApi()

def http_error_handler(status, message, traceback, version):
  return jsonHttpErrorResponse(message)

def error_handler():
  exceptionInfo = exc_info()
  excType = exceptionInfo[0]
  exception = exceptionInfo[1]

  msg = str(exception) 

  logging.warning("error_handler exception type: " + str(excType.__name__) + " message: " + str(exception))

  return jsonHttpErrorResponse(msg)

cherrypy.config.update({
    'request.error_response': error_handler,
    'request.error_page': {"default": http_error_handler},
    'server.socket_host': '0.0.0.0',
    'server.socket_port': args.port,
    'log.screen': False,
    'tools.encode.on': True,
    'tools.encode.encoding': 'utf-8'
});

logging.getLogger().setLevel(args.logging)
logging.basicConfig(format = "[%(levelname)s] %(message)s ")

config = {}

setCookieFilename(args.cookieFilename)
setupCredentialsJson()

cherrypy.quickstart(FrontendWrapper(), '/', config)
