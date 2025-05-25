"""
An orgviz "model" is the logical structure of the graph, that is passed to 
GraphViz.
"""

from orgviz.person import Person

class Model():
    """
    Defines a logical model.
    """

    def __init__(self):
        self.title = "Untitled Organization"
        self.people = {}
        self.edges = []
        self.teams = set()
        self.lastPerson = None

    def findPerson(self, name):
        for person in self.people.values():
            if Person.getDotNodeNameFromFullName(person.getAttribute('id')) == name:
                return person
        return self.getPersonByName(name)

    def getPersonByName(self, personFullName):
        if personFullName in self.people:
            return self.people[personFullName]

        raise RuntimeError("Person not found: " + personFullName)

    def addPerson(self, personFullName):
        person = Person(personFullName)

        self.people[person.dotNodeName] = person
        self.lastPerson = person

    def addTeam(self, team):
        self.teams.add(team)

    def addConnection(self, edgeType, destination):
        self.edges.append({
            "origin": self.lastPerson.dotNodeName,
            "type": edgeType,
            "destination": Person.getDotNodeNameFromFullName(destination) 
        })

# pylint: disable=too-few-public-methods,too-many-instance-attributes
# ^^ Because we need to hold lots of options somehow. namedtuple would be too 
# much to maintain. Can't think of a "better" way, or a problem with this 
class ModelOptions():
    """
    Options used for drawing/building the model.
    """

    def __init__(self):
        self.skipDrawingLegend = False
        self.skipDrawingTitle = False
        self.skipDrawingTeams = False
        self.vizType = "DS"
        self.outputType = "svg"
        self.teams = []
        self.attributeMatches = []
        self.influence = []
        self.profilePictures = False
        self.profilePictureDirectory = "/opt/"
        self.dpi = 100

