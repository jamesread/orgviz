#!/usr/bin/env python3

import cherrypy
import requests

import configargparse
import os
import json

import logging
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "orgviz"))

from sys import exc_info

from orgviz.modelToGraphviz import ModelToGraphVizConverter
from orgviz.model import ModelOptions
from orgviz.modelParser import parseModel
from orgviz.graphviz import runDot

parser = configargparse.ArgumentParser(default_config_files = ["~/.orgviz-web.cfg"])
parser.add_argument("--logging", type = int, default = 20, help = "1 = Everything. 50 = Critical only.", env_var = "LOGGING")
parser.add_argument('--port', default = 8081, type = int, env_var = "PORT");
parser.add_argument('--outputDirectoryLocal', default = "/var/www/html/orgvizOutput/", env_var = "OUTPUT_DIRECTORY_LOCAL");
parser.add_argument('--outputDirectoryPublic', default = "http://localhost:8081/output/", env_var = "OUTPUT_DIRECTORY_PUBLIC");
parser.add_argument('--serverMode', default = 'directInput', choices = ['directInput', 'webservice'], env_var = "SERVER_MODE");
parser.add_argument('--webserviceUrl', default = 'http://localhost', env_var = "WEBSERVICE_URL");
parser.add_argument('--webserviceName', default = 'From Webservice', env_var = "WEBSERVICE_NAME");
parser.add_argument('--webserviceKeyName', default = 'Key', env_var = "WEBSERVICE_KEY_NAME");
args = parser.parse_args();

def jsonHttpErrorResponse(errors, filename = ""):
    cherrypy.response.headers['Content-Type'] = 'application/json'
    cherrypy.response.status = 400;

    return json.dumps({
        "errors": errors,
        "filename": filename
    }, ensure_ascii=False).encode('utf8')

class FrontendWrapper:
    @cherrypy.expose
    def index(self):
        return '<h1>orgviz API!</h1><p>Hi! Did you want the <a href = "webui">webui</a>?'

    @cherrypy.expose
    @cherrypy.tools.json_out()
    def createImageFromOrgViz(self):
        errors = list();
        body = cherrypy.request.body.read().decode('utf-8')
        dotOutput = False

        if len(body) == 0:
            errors.append("Error: No content")
        else:
            opts = ModelOptions()
            opts.skipDrawingLegend = True
            opts.skipDrawingTitle = True

            conv = ModelToGraphVizConverter(opts=opts)
            mdl = parseModel(body)

            try: 
                body = conv.getModelAsDot(mdl)
            except Exception as e:
                errors.append(str(e))

            dotOutput = self.orgStringToDot(body, errors)

            return self.dotReturn(dotOutput, errors)


    def dotReturn(self, dotResult, errors):
        if len(errors) != 0:
            return {
                "errors": errors,
                "filename": None
            }

        if dotResult:
            logging.info("Dot ran okay")

            return {
                "errors": list(),
                "filename": args.outputDirectoryPublic + "orgviz.png" 
            }
        else:
            logging.info("Dot failure: " + str(errors))

            errors.append("Dot failure")
            
            return {
                "errors": errors,
                "filename": None
            }


    def orgStringToDot(self, body, errors):
        f = open("/tmp/orgfile", "w")
        f.write(body);
        f.close()

        dotOutput, error = runDot("/tmp/orgfile", args.outputDirectoryLocal + "/orgviz.png")

        if error != None:
            logging.info("Adding Dot error to list")

            errors.append(error)

        return dotOutput


    @cherrypy.expose
    @cherrypy.tools.json_out()
    def createImageFromWebservice(self):
        key = cherrypy.request.body.read().decode('utf-8')

        logging.info("createImageFromWebservice, Key is " + key) 
        logging.info("createImageFromWebservice, Webservice Base URL is " + args.webserviceUrl) 

        url = args.webserviceUrl + "/" + key 

        logging.info("createImageFromWebservice, Webservice Full URL is " + url)

        r = requests.get(url = url, timeout = 30)
        
        logging.info("createImageFromWebservice, HTTP res is " + str(r.status_code)) 

        errors = list()

        if (r.status_code == 200):
            orgFileFromWebservice = r.content.decode('utf-8')

            opts = ModelOptions()
            opts.skipDrawingLegend = True
            opts.skipDrawingTitle = True

            conv = ModelToGraphVizConverter(opts=opts)
            mdl = parseModel(orgFileFromWebservice)

            try: 
                body = conv.getModelAsDot(mdl)
            except Exception as e:
                errors.append(str(e))

            dotOutput = self.orgStringToDot(body, errors)
        else:
            try: 
                errorOutput = r.content.decode('utf-8')

                logging.info("createImageFromWebservicem, errorOutput is: " + str(errorOutput))

                errorJson = r.json()

                if "errorMessage" in errorJson:
                    errors.append(errorJson["errorMessage"])
                else:
                    errors.append("Upstream status code is: " + str(r.status_code))
            except:
                errors.append("Upstream status code is: " + str(r.status_code) + ", and could not determine the error message either!")

            dotOutput = ""

        return self.dotReturn(dotOutput, errors)


    @cherrypy.expose
    @cherrypy.tools.json_out()
    def clientConfig(self):
        return {
            "serverMode": args.serverMode,
            "webserviceName": args.webserviceName,
            "webserviceKeyName": args.webserviceKeyName,
        }

def http_error_handler(status, message, traceback, version):
  return jsonHttpErrorResponse([message])

def error_handler():
  exceptionInfo = exc_info()
  excType = exceptionInfo[0]
  exception = exceptionInfo[1]

  msg = str(exception)
  
  logging.warning("error_handler exception type: " + str(excType.__name__) + ". message: " + str(exception))

  cherrypy.response.body = jsonHttpErrorResponse([msg])

cherrypy.config.update({
    'server.socket_host': '0.0.0.0',
    'server.socket_port': args.port,
    'log.screen': False
});

config = {
     '/' : {
        'request.error_response': error_handler,
        'request.error_page': {"default": http_error_handler}
    },
   '/webui': {
    'tools.staticdir.on': True,
    'tools.staticdir.dir': 'webui/dist',
    'tools.staticdir.root': os.path.dirname(os.path.realpath(__file__)),
    'tools.staticdir.index': 'index.html'
    },
    '/output': {
        'tools.staticdir.on': True,
        'tools.staticdir.dir': args.outputDirectoryLocal,
        'tools.expires.on': True,
        'tools.expires.secs': 1,
    }
}

logging.getLogger().setLevel(args.logging)
logging.basicConfig(format = "[%(levelname)s] %(message)s ")

cherrypy.quickstart(FrontendWrapper(), '/', config)

