"""
Parses orgviz models from strings into model.Model
"""

import logging
from orgviz.model import Model

def parseModel(contents):
    model = Model()

    for line in contents.split("\n"):
        if line == "" or line.startswith('#'): 
            continue

        if line.startswith("@") and ":" in line:
            line = line.replace("@", "")
            key, val = line.split(":", 1)

            if key == "title": 
                model.title = val.strip()

            continue

        if line.startswith("\t"):
            line = line.strip().replace("\t", "")

            if "->" in line:
                parseConnection(model, line)
                continue

            if ":" in line:
                parsePersonProperty(model, line)
                continue

            logging.warning("Cannot parse line: %s", line)


        model.addPerson(line)

    return model

def parseConnection(model, line):
    edgeType, destination = line.split("->", 1)

    if edgeType != "" and destination != "":
        model.addConnection(edgeType.strip(), destination.strip())

def parsePersonProperty(model, line):
    propertyKey, propertyValue = line.split(":", 1)
    propertyKey = propertyKey.strip().lower()
    propertyValue = propertyValue.strip()

    if propertyKey == "influence":
        model.lastPerson.setInfluence(propertyValue)
        return

    if propertyKey == "sentiment":
        model.lastPerson.setSentiment(propertyValue)
        return

    if propertyKey == "dmu":
        model.lastPerson.setDmu(propertyValue)
        return

    if propertyKey == "team":
        model.addTeam(propertyValue)
        model.lastPerson.setTeam(propertyValue)
        return

    model.lastPerson.setAttribute(propertyKey, propertyValue)

