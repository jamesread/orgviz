import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import { ClientToFrontend } from './gen/proto/orgviz/clientapi/v1/clientapi_pb';

import * as d3 from 'd3';
import { OrgChart } from 'd3-org-chart';

export function init() {
  createApiClient();
  setupApi()

  initChart();
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

	document.getElementById('current-version').innerText = 'Version: ' + status.version;
}

async function initChart() {
  document.getElementById('chart').innerHTML = ''; // Clear previous chart

  const data = await window.client.getChart({});
  let people = data.people;

  people[0].parentId = null; // Ensure the root node has no parent

  window.people = people

  window.chart = new OrgChart().container('#chart');
  window.chart.nodeContent(renderNodeContent)

  window.chart.data(people).render();
}

function renderNodeContent(d) {
  return `<div class="person">
            <h3>${d.data.fullName}</h3>
            <p>ID: ${d.data.id}</p>
          </div>`;
}
