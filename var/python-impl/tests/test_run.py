import unittest

from orgviz.graphviz import runDot

class RunnerTestCase(unittest.TestCase):
    def test_influence(self):
        runDot("tests/badOrgFiles/emptyAttributeValues.org", "/tmp/orgviz.png")


