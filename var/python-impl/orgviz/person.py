"""
Contains a class that represents a logical Person. 
"""

import logging

class Person():
    """
    A person.
    """

    def __init__(self, fullName):
        self.fullName = fullName.strip()

        self.dotNodeName = Person.getDotNodeNameFromFullName(self.fullName)
        self.team = "??"
        self.influence = ""
        self.dmu = "U"
        self.sentiment = "N"
        self.attributes = {}

    @staticmethod
    def getDotNodeNameFromFullName(fullName):
        ret = fullName.strip().replace("\"", "\\\"")
        ret = "\"" + ret + "\""

        return ret

    def getDmuDescription(self):
        return {
            "D": "Decision Maker",
            "B": "Buyer",
            "I": "Influencer",
            "G": "Gatekeeper",
            "U": "User",
        }.get(self.dmu, "dmu?")

    def getSentimentDescription(self):
        return {
            "P": "Proponent",
            "N": "Neutral",
            "O": "Opponent"
        }.get(self.sentiment, "sentiment?")

    def setDmu(self, newValue):
        dmu = newValue.upper().strip()

        if dmu in ["D", "DECISION MAKER", "DM"]: 
            self.dmu = "D"
            return

        if dmu in ["B", "BUYER", "BUY"]:
            self.dmu = "B"
            return

        if dmu in ["I", "INFLUENCER"]:
            self.dmu = "I"
            return

        if dmu in ["G", "GATEKEEPER"]:
            self.dmu = "G"
            return

        if dmu in ["U", "USER"]:
            self.dmu = "U"
            return

        logging.warning("Unknown `dmu` (decision making unit) for: %s, got: %s, but should be [D]ecision Maker, [B]uyer, [I]nfluencer, [G]atekeeper, [U]ser", self.fullName, dmu)
        self.dmu = "U"

    def setSentiment(self, newValue):
        sentiment = newValue.upper().strip()

        if sentiment in ["P", "PROMOTOR", "PROMOTER", "PROPONENT"]:
            self.sentiment = "P"
            return

        if sentiment in ["O", "OPPONENT"]:
            self.sentiment = "O"
            return

        if sentiment in ["N", "NEUTRAL"]:
            self.sentiment = "N"
            return

        logging.warning("Unknown `sentiment` for: %s, got %s, but should be [P]romoter, [O]pponent or [N]eutral", self.fullName, sentiment)
        self.sentiment = "N"

    def setInfluence(self, influence):
        self.influence = influence.strip()

    def setTeam(self, team):
        self.team = team

    def hasAttribute(self, key):
        return key in self.attributes

    def getAttribute(self, key):
        if key in self.attributes:
            rawValue = self.attributes[key]

            safeValue = rawValue.replace('&', '&amp;')
            safeValue = safeValue.replace('|', ' ')
            safeValue = safeValue.strip()

            if safeValue != "": 
                return safeValue

        return "Unknown: " + key

    def setAttribute(self, key, value):
        self.attributes[key] = value

