#!/usr/bin/env python3

import sys
import logging
import tempfile
import os

from os import listdir
from os.path import isfile, join
from configargparse import ArgParser

from orgviz.modelToGraphviz import ModelToGraphVizConverter
from orgviz.modelParser import parseModel
from orgviz.model import ModelOptions
from orgviz.graphviz import runDot

sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "orgviz"))

args = None

def getArgumentParser():
    parser = ArgParser(default_config_files = ["~/.orgviz.cfg"])
    parser.add_argument("--input", "-I", default = None, env_var = "ORGVIZ_INPUT")
    parser.add_argument("--output", "-O", default = os.getcwd())
    parser.add_argument("--skipDrawingLegend", "-L", action = "store_true")
    parser.add_argument("--skipDrawingTeams", action = "store_true")
    parser.add_argument("--skipDrawingTitle", action = "store_true")
    parser.add_argument("--dotout", action = "store_true")
    parser.add_argument("--logging", type = int, default = 20, help = "1 = Everything. 50 = Critical only.")
    parser.add_argument("--teams", nargs = "*", default = [])
    parser.add_argument("--influence", nargs = "*", default = [], choices = ["supporter", "promoter", "enemy", "internal"])
    parser.add_argument("--profilePictureDirectory", default = "/opt/profilePictures/", help = "A directory containing [name].jpeg files of people in your organization.")
    parser.add_argument("--profilePictures", "-P", action = "store_true")
    parser.add_argument("--outputType", "-T", default = "svg", choices = ["png", "svg"])
    parser.add_argument("--keepDotfile", action = "store_false")
    parser.add_argument("--vizType", choices = ["DS", "inf", "none"], default = "DS")
    parser.add_argument("--dpi", type = int, default = 100, help = 'DPI (resolution), only used for PNG.')
    parser.add_argument("--attributeMatches", "-a", nargs = "*", default = [], metavar = "KEY=VALUE")

    return parser

def findFirstOrgfileInCurrentDirectory():
    cwd = os.getcwd()

    files = [f for f in listdir(cwd) if isfile(join(cwd, f))]

    for f in files:
        if f.endswith(".org"):
            return f

    logging.warning("Could not find a .org file in this directory.")
    return "?"

def getFileContents():
    if args.input is None:
        args.input = findFirstOrgfileInCurrentDirectory()

    logging.info("Reading org file: %s", args.input)

    try: 
        with open(args.input, 'r', encoding='utf8') as f:
            contents = f.read()

        return contents
    except FileNotFoundError:
        logging.fatal("Could not find input file: %s", args.input)
        sys.exit()

def main():
    global args

    args = getArgumentParser().parse_args()

    logging.getLogger().setLevel(args.logging)
    logging.basicConfig(format = "[%(levelname)s] %(message)s ")

    logging.debug("Teams: %s", str(args.teams))

    opts = ModelOptions()
    opts.skipDrawingLegend = args.skipDrawingLegend
    opts.skipDrawingTitle = args.skipDrawingTitle
    opts.skipDrawingTeams = args.skipDrawingTeams
    opts.vizType = args.vizType
    opts.outputType = args.outputType
    opts.teams = args.teams
    opts.attributeMatches = args.attributeMatches
    opts.influence = args.influence
    opts.profilePictures = args.profilePictures
    opts.profilePictureDirectory = args.profilePictureDirectory
    opts.dpi = args.dpi

    out = None

    try: 
        mdl = parseModel(getFileContents())
        converter = ModelToGraphVizConverter(opts)

        out = converter.getModelAsDot(mdl)
    except Exception as exception:
        logging.error(str(exception))


    if args.dotout:
        logging.warning("Outputting graph to stdout, not running Dot")

        print(out)
    elif out is not None:
        with tempfile.NamedTemporaryFile(delete = args.keepDotfile) as temporaryGraphvizFile:
            temporaryGraphvizFile.write(bytes(out, "UTF-8"))
            temporaryGraphvizFile.flush()

        logging.debug("Wrote DOT file to: %s", str(temporaryGraphvizFile.name))

        outputImageFilename = args.output + "/orgviz." + args.outputType
        
        runDot(str(temporaryGraphvizFile.name), outputImageFilename, args.outputType)

        temporaryGraphvizFile.close() # will delege the org file if !args.keepDotFile

if __name__ == "__main__":
    main()

