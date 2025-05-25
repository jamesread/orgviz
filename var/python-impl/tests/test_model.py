import unittest

from orgviz.model import Model, ModelOptions
from orgviz.person import Person

class ModelTestCase(unittest.TestCase):
    def test_add_and_find_person(self):
        m = Model()
        
        assert len(m.people) == 0

        m.addPerson("Dave")

        assert len(m.people) == 1

        with self.assertRaises(Exception):
            m.findPerson(Person.getDotNodeNameFromFullName("Gwendelina"))

        assert m.findPerson(Person.getDotNodeNameFromFullName("Dave")) is not None

    def test_find_by_id(self):
        m = Model()

        m.addPerson("Dave")
        m.lastPerson.setAttribute("id", "superman")

        self.assertTrue(m.findPerson(Person.getDotNodeNameFromFullName("Dave")) is not None)
        self.assertTrue(m.findPerson(Person.getDotNodeNameFromFullName("superman")) is not None)

        with self.assertRaises(Exception):
            m.findPerson(Person.getDotNodeNameFromFullName("superted"))


    def test_add_team(self):
        m = Model()
        m.addTeam("Team 1")
        m.addTeam("Team 2")

        self.assertIn("Team 1", m.teams)
        self.assertIn("Team 2", m.teams)
        self.assertNotIn("Team 3", m.teams)

    def test_add_connection(self):
        m = Model();

        self.assertTrue(len(m.edges) == 0)

        m.addPerson("Alice")
        m.addPerson("Bob")
        m.addConnection("loves", "Alice")

        self.assertTrue(len(m.edges) == 1)

    def test_create_model_options(self):
        mo = ModelOptions()
