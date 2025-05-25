import unittest

from orgviz.model import ModelOptions
from orgviz.modelToGraphviz import ModelToGraphVizConverter
from orgviz.modelParser import parseModel

class ParserTestCase(unittest.TestCase):
    def readOrgFile(self, filename):
        f = open(filename, 'r', encoding='utf8')
        contents = f.read()
        f.close()

        return contents

    def test_unknown_attributes(self):
        mdl = parseModel(self.readOrgFile("tests/badOrgFiles/unknownAttributes.org"))

        self.assertTrue(mdl is not None);

    def test_lines_no_colon(self):
        mdl = parseModel(self.readOrgFile("tests/badOrgFiles/linesNoColon.org"))

        self.assertTrue(mdl is not None);

    def test_parse_and_generate_example_org(self):
        opts = ModelOptions()
        conv = ModelToGraphVizConverter(opts)

        mdl = parseModel(self.readOrgFile("examples/ExampleCompany.org"))

        self.assertTrue(mdl is not None);
 
        dot = conv.getModelAsDot(mdl)

        self.assertTrue(dot is not None)

    def test_influence(self):
        opts = ModelOptions()
        opts.vizType = "inf"

        mdl = parseModel(self.readOrgFile("tests/badOrgFiles/simpleInfluence.org"))

        conv = ModelToGraphVizConverter(opts)
        conv.getModelAsDot(mdl)

    def test_influence(self):
        opts = ModelOptions()

        mdl = parseModel(self.readOrgFile("tests/badOrgFiles/emptyAttributeValues.org"))

        conv = ModelToGraphVizConverter(opts)
        conv.getModelAsDot(mdl)

