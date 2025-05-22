"""
Contains on real class, ModelToGraphVizConverter, which, aherm, converts orgviz
models to GraphViz. 
"""

import logging
import os

from orgviz.model import ModelOptions

class ModelToGraphVizConverter():
    """
    Converts orgviz models into Graphviz output.
    """

    def __init__(self, opts: ModelOptions):
        self.opts = opts

    def getInfluenceStyleAsDot(self, influence):
        if self.opts.vizType != "inf": 
            return "fillcolor=white, style=filled"

        vizTypeStyles = {
            "supporter": "fillcolor=skyblue, style=filled",
            "promoter": "fillcolor=GreenYellow, style=filled",
            "enemy": "fillcolor=salmon, style=filled",
            "internal": "fillcolor=black, style=filled, fontcolor=white",
            "": "fillcolor=white, style=filled",
        }

        if influence in vizTypeStyles: 
            return vizTypeStyles[influence]

        logging.warning("Unknown influence class: %s", influence)
        return ""

    def getLegendAsDot(self):
        out = ""

        if not self.opts.skipDrawingLegend:
            if self.opts.vizType == "DS":
                pass
            
            if self.opts.vizType == "inf":
                out += "subgraph cluster_00 {\n"
                out += "label=Legend\n"
                out += "fillcolor=beige\n"
                out += "style=filled\n"
                out += "node [fontsize=9]\n"
                out += "supporter [fillcolor=skyblue, style=filled]\n"
                out += "promoter [fillcolor=GreenYellow, style=filled]\n"
                out += "hostile [fillcolor=salmon, style=filled]\n"
                out += "internal [fillcolor=black, fontcolor = white, style=filled]\n"
                out += "}\n"

        return out

    def isPersonExcluded(self, person):
        if len(self.opts.attributeMatches) > 0: 
            for attributeSearch in self.opts.attributeMatches:
                if "=" not in attributeSearch: 
                    continue

                key, val = map(lambda i: i.strip(), attributeSearch.split("=", 1))

                print("key, val", key, val, person.getAttribute(key))

                if val not in person.getAttribute(key):
                    return True

        if len(self.opts.teams) > 0 and person.team not in self.opts.teams:
            return True

        if len(self.opts.influence) > 0 and person.influence not in self.opts.influence:
            return True

        return False

    def getTeamsAsDot(self, model):
        if self.opts.skipDrawingTeams:
            return ""

        out = ""

        subGraphCount = 0

        # A useful behavior of dot, is that is a subgraph is empty (IE, no nodes), 
        # then it's rendering is skipped. This is why we blindly define all teams 
        # (subgraphs), and only check if people are excluded. 

        for teamName in model.teams:
            subGraphCount += 1

            out += "subgraph cluster_" + str(subGraphCount) + "{\n"
            out += 'label="' + teamName + '"' + "\n"
            out += "style=filled\n"
            out += "labelloc=b\n" # Because the graph is "drawn upside down", b -> t
            out += "fillcolor=skyblue\n"

            for person in model.people.values():
                if self.isPersonExcluded(person): 
                    continue

                if person.team == teamName:
                    out += person.dotNodeName + " []\n"

            out += "}\n"

        return out

    # ^^ Is logically part of this class still.
    def getEdgeDotStyle(self, edge):
        if edge['type'] == "supports":
            return "style=dotted"

        return ""

    # ^^ Is logically part of this class still.
    def getStyleForDmu(self, dmu):
        return {
            "D": "greenyellow",
            "I": "skyblue",
            "B": "orchid",
            "G": "salmon",
            "U": "gray"
        }.get(dmu, "white")

    # ^^ Is logically part of this class still.
    def getStyleForSentiment(self, sentiment):
        return {
            "P": "greenyellow",
            "N": "yellow",
            "O": "salmon",
        }.get(sentiment, "white")

    # pylint: disable=too-many-function-args
    # ^^ seems bogus here
    def getDsVisType(self, person):
        ret = f'<tr><td bgcolor = "{self.getStyleForDmu(person.dmu)}" border = "1">{person.getDmuDescription()}</td><td bgcolor = "{self.getStyleForSentiment(person.sentiment)}" border = "1">{person.getSentimentDescription()}</td></tr>'

        return ret

    def getPersonLabelAsDot(self, person):
        ret = '<<table border = "0" cellspacing = "0">'
        ret += f'<tr><td border = "1" colspan = "2"><b>{person.fullName}</b></td></tr>'
        ret += f'<tr><td border = "1" colspan = "2"><font point-size = "9">{person.getAttribute("title")}</font></td></tr>'

        if self.opts.profilePictures:
            pic = self.opts.profilePictureDirectory + person.fullName + ".jpeg"

            if os.path.exists(pic):
                logging.debug("Found LinkedIn profile: %s", pic)

                ret += f'<tr><td colspan = "2" border = "1"><img src = "{pic}" /></td></tr>'
            else:
                logging.warning("No profile pic found for %s", pic)

        if person.hasAttribute("country"):
            ret += '<tr><td colspan = "2">' + person.getAttribute('country')  +  '</td></tr>'

        if self.opts.vizType == "DS":
            ret += self.getDsVisType(person)

        ret += "</table>>"
        return ret


    def getModelAsDot(self, model):
        out = ""
        out += "digraph {\n"
        out += "rankdir = BT;\n"

        if self.opts.outputType == "png":
            out += "graph [ dpi = " + str(self.opts.dpi) +  " ]\n"

        if not self.opts.skipDrawingTitle:
            out += 'label="' + model.title + ' - github.com/jamesread/orgviz"' + "\n"

        out += 'labelloc="t"' + "\n"
        out += 'fontname=Overpass' + "\n"
        out += "node [fontname=Overpass, shape=record]\n"
        out += "edge [fontname=Overpass, fontsize=9]\n"

        out += self.getTeamsAsDot(model)

        for person in model.people.values():
            if self.isPersonExcluded(person): 
                continue

            out += f"{person.dotNodeName} [margin=0, border=invisible, label={self.getPersonLabelAsDot(person)},{self.getInfluenceStyleAsDot(person.influence)}]\n"

        for edge in model.edges:
            if self.isPersonExcluded(model.findPerson(edge['origin'])) or self.isPersonExcluded(model.findPerson(edge['destination'])): 
                continue

            out += f'{edge["origin"]} -> {model.findPerson(edge["destination"]).dotNodeName} [label="{edge["type"]}", {self.getEdgeDotStyle(edge)}]' + "\n"

        out += self.getLegendAsDot()
        out += "}"
        
        return out


