package clientapi

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/jamesread/golure/pkg/dirs"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
	"github.com/jamesread/orgviz/internal/buildinfo"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type clientApi struct {
	orgfiles []*OrgFile
}

type OrgFile struct {
	Title string
	People []*Person

	peopleIds map[string]int32
}

type Person struct {
	Name string
	Alias string
	Title string
	Team string
	Reports string
	Attributes map[string]string

	id int32
	reportsToId int32
}

func NewServer() *clientApi {
	api := &clientApi{}
	api.orgfiles = make([]*OrgFile, 0)

	api.readOrgFiles()

	return api
}

func (self *clientApi) ReadOrgFile(filename string) {
	contents, err := os.ReadFile(filename)

	if err != nil {
		log.Errorf("Error opening org file %s: %v", filename, err)
		return
	}

	parsed := &OrgFile{}

	err = yaml.Unmarshal(contents, parsed)

	if err != nil {
		log.Errorf("Error parsing org file %s: %v", filename, err)
		return
	}

	log.Infof("Read org file: %s, title: %+v", filename, parsed)

	buildPeopleIds(parsed)
	linkPeopleIds(parsed)

	self.orgfiles = append(self.orgfiles, parsed)
}

func buildPeopleIds(orgfile *OrgFile) {
	orgfile.peopleIds = make(map[string]int32)

	log.Infof("People in org file: %v", len(orgfile.People))

	for _, person := range orgfile.People {
		name := person.Name

		_, found := orgfile.peopleIds[name]

		if !found {
			newId := int32(len(orgfile.peopleIds) + 1)

			orgfile.peopleIds[name] = newId

			if person.Alias != "" {
				orgfile.peopleIds[person.Alias] = newId
			}
		}

		person.id = orgfile.peopleIds[name]
	}

	log.Infof("Parsed org file: %+v, len %v", orgfile, len(orgfile.peopleIds))
}

func linkPeopleIds(orgfile *OrgFile) {
	for _, person := range orgfile.People {
		if person.Reports != "" {
			person.reportsToId = orgfile.peopleIds[person.Reports]
		}
	}
}

func (api *clientApi) readOrgFiles() {
	orgdir, _ := dirs.GetFirstExistingDirectory([]string{
		"/config/orgs/",
		"../examples/",
	});

	log.Infof("Reading org files from directory: %s", orgdir)

	files, _ := filepath.Glob(orgdir + "/*.yml")

	for _, file := range files {
		log.Infof("Reading org file: %v", file)

		api.ReadOrgFile(file)
	}
	/*
	api.AddPerson(1, -1, "James Read", "Benevolent Dictator For Life")
	api.AddPerson(2, 1, "Alice Smith", "Product Manager")
	api.AddPerson(3, 1, "Bob Johnson", "Designer")
	api.AddPerson(4, 2, "Charles Brown", "Software Engineer")
	api.AddPerson(5, 2, "Dave Prince", "QA Engineer")
	api.AddPerson(6, 3, "Eve Adams", "UX Researcher")
	api.AddPerson(7, 4, "Frank Castle", "DevOps Engineer")
	api.AddPerson(8, 5, "Grace Hopper", "Data Scientist")
	api.AddPerson(9, 6, "Hank Pym", "Security Analyst")
	*/
}

func (c *clientApi) GetClientInitialSettings(ctx context.Context, req *connect.Request[pb.GetClientInitialSettingsRequest]) (*connect.Response[pb.GetClientInitialSettingsResponse], error) {
	// This is where you would implement the logic to retrieve the initial settings for the client.
	// For now, we return an empty response.
	response := &pb.GetClientInitialSettingsResponse{
		Version: buildinfo.Version,
	}

	for idx, orgfile := range c.orgfiles {
		response.Charts = append(response.Charts, &pb.ChartInfo{
			ChartId: strconv.Itoa(idx),
			Title: orgfile.Title,
		})
	}

	return connect.NewResponse(response), nil

}

func chartPeopleToPbPeople(people []*Person) []*pb.Person {
	pbPeople := make([]*pb.Person, len(people))

	for i, person := range people {
		pbPeople[i] = &pb.Person{
			FullName:        person.Name,
			JobTitle:       person.Title,
			Id:			 person.id,
			ParentId:	  person.reportsToId,
			AvatarUrl:    findAvatarUrl(person),
		}
	}

	return pbPeople
}

func findAvatarUrl(person *Person) string {
	nameBits := strings.Split(person.Name, " ")

	if len(nameBits) > 0 {
		firstName := nameBits[0]
		firstName = strings.ToLower(firstName)

		return "/avatars/" + firstName + ".jpg"
	}

	return ""
}

func (c *clientApi) GetChart(ctx context.Context, req *connect.Request[pb.GetChartRequest]) (*connect.Response[pb.GetChartResponse], error) {
	chartId, _ := strconv.Atoi(req.Msg.ChartId)
	chart := c.orgfiles[chartId]

	response := &pb.GetChartResponse{
		Title: chart.Title,
		People: chartPeopleToPbPeople(chart.People),
	}

	return connect.NewResponse(response), nil
}

//func (c *ClientApi) Init(ctx context.Context, req *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
