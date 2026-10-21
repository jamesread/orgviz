package clientapi

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"github.com/jamesread/golure/pkg/dirs"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
	"github.com/jamesread/orgviz/internal/buildinfo"
	"github.com/jamesread/orgviz/internal/config"
	"github.com/jamesread/orgviz/internal/orgs"
	log "github.com/sirupsen/logrus"
)

type clientAPI struct {
	cfg      *config.Config
	orgfiles []*orgs.File
}

// NewServer loads organisation files and serves the client API.
func NewServer(cfg *config.Config) *clientAPI {
	api := &clientAPI{cfg: cfg}
	api.orgfiles = loadOrgFiles(cfg)
	return api
}

func loadOrgFiles(cfg *config.Config) []*orgs.File {
	orgDir := findOrgsDir(cfg)
	log.WithField("dir", orgDir).Info("Reading org files")

	loaded, err := orgs.LoadDir(orgDir)
	if err != nil {
		log.WithError(err).Error("Failed to load org files")
		return []*orgs.File{}
	}

	return loaded
}

func findOrgsDir(cfg *config.Config) string {
	candidates := []string{
		"../var/config-skel/orgs/",
		"/config/orgs/",
		"../examples/",
	}

	if cfg != nil && cfg.OrgsDir() != "" {
		candidates = append([]string{cfg.OrgsDir()}, candidates...)
	}

	orgDir, err := dirs.GetFirstExistingDirectory("orgs", candidates)
	if err != nil {
		log.WithError(err).Warn("No orgs directory found")
		return "not-found"
	}

	return orgDir
}

func (c *clientAPI) GetClientInitialSettings(ctx context.Context, req *connect.Request[pb.GetClientInitialSettingsRequest]) (*connect.Response[pb.GetClientInitialSettingsResponse], error) {
	response := &pb.GetClientInitialSettingsResponse{
		Version: buildinfo.Version,
	}

	for idx, orgfile := range c.orgfiles {
		response.Charts = append(response.Charts, &pb.ChartInfo{
			ChartId: strconv.Itoa(idx),
			Title:   orgfile.Title,
		})
	}

	return connect.NewResponse(response), nil
}

func (c *clientAPI) GetChart(ctx context.Context, req *connect.Request[pb.GetChartRequest]) (*connect.Response[pb.GetChartResponse], error) {
	chart, err := c.chartByID(req.Msg.ChartId)
	if err != nil {
		return nil, err
	}

	response := &pb.GetChartResponse{
		ChartId: req.Msg.ChartId,
		Title:   chart.Title,
		People:  chartPeopleToPb(chart.People),
	}

	return connect.NewResponse(response), nil
}

func (c *clientAPI) chartByID(chartID string) (*orgs.File, error) {
	idx, err := strconv.Atoi(chartID)
	if err != nil || idx < 0 || idx >= len(c.orgfiles) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown chart %q", chartID))
	}

	return c.orgfiles[idx], nil
}

func chartPeopleToPb(people []*orgs.Person) []*pb.Person {
	pbPeople := make([]*pb.Person, len(people))

	for i, person := range people {
		pbPeople[i] = &pb.Person{
			FullName:  person.Name,
			JobTitle:  person.Title,
			Id:        person.ID,
			ParentId:  person.ReportsToID,
			AvatarUrl: orgs.AvatarURL(person.Name),
		}
	}

	return pbPeople
}
