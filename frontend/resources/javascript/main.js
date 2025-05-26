import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import { ClientToFrontend } from './gen/proto/orgviz/clientapi/v1/clientapi_pb';

import * as d3 from 'd3';
import { OrgChart } from 'd3-org-chart';

export function init() {
  createApiClient();
  setupApi()

  document.getElementById('change-layout').onclick = () => {
    changeLayout()
  }
  console.log('Initializing OrgViz...');

  window.isCompact = true;

  initChart();
  loadChart('0')
}

function changeLayout() {
  document.getElementById('change-layout').innerText = window.isCompact ? 'Compact layout' : 'Expand layout';


  window.isCompact = !window.isCompact;

  console.log('Changing layout to', window.isCompact ? 'compact' : 'expanded');

  window.chart.compact(window.isCompact).render().fit()
}

function createApiClient() {
	let baseUrl = '/api/'

	if (window.location.hostname.includes('localhost')) {
		baseUrl = 'http://localhost:8080/api/'
	}

	window.transport = createConnectTransport({
		baseUrl: baseUrl,
	})

	window.client = createClient(ClientToFrontend, window.transport)
}

async function setupApi() {
  const status = await window.client.getClientInitialSettings()

  const chartsList = document.getElementById('charts-list');

  for (let chart of status.charts) {
    const href = document.createElement('a');
    href.innerText = chart.title;
    href.href = '#';
    href.onclick = () => {
      loadChart(chart.chartId);
    }

    const li = document.createElement('li');
    li.appendChild(href)

    chartsList.appendChild(li);
  }

  document.getElementById('current-version').innerText = 'Version: ' + status.version;
}

function initChart() {
}

async function loadChart(idToLoad) {
  console.log('Loading chart with ID:', idToLoad);

  document.getElementById('chart').innerHTML = ''; // Clear previous chart

  const data = await window.client.getChart({
    chartId: idToLoad,
  });

  let people = data.people;

  // Clean people reporting lines
  for (let i = 0; i < people.length; i++) {
    if (people[i].parentId === 0) {
      people[i].parentId = 1;
    }
  }
  people[0].parentId = null;

  // Set avatar url
  for (let person of people) {
    if (person.avatarUrl === '') {
      person.avatarUrl = '/avatars/default.png';
    }
  }

  window.people = people

  window.chart = new OrgChart()
    .nodeHeight((d) => 85)
    .container('#chart')
    .childrenMargin((d) => 50)
    .compactMarginBetween((d) => 35)
    .compactMarginPair((d) => 30)
    .neighbourMargin((a, b) => 20)
    .compact(window.isCompact)
    .linkUpdate(function(d, i, arr) {
      return d3.select(this).attr('stroke', '#666')
    })

  window.chart.nodeContent(renderNodeContent)

  window.chart.data(people).render();
}

function renderNodeContent(d) {
  return `<div class="person">
            <img src="${d.data.avatarUrl}" alt="${d.data.fullName}" class="avatar" />
            <div class="person-info">
            <h3>${d.data.fullName}</h3>
            <p>${d.data.jobTitle}</p>
            </div>

          </div>`;
}
