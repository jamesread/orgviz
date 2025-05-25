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

function initChart() {
  document.getElementById('chart').innerHTML = ''; // Clear previous chart

  const data = [
    { id: 1, parentId: null, name: 'Alice' },
    { id: 2, parentId: 1, name: 'Bob' },
    { id: 3, parentId: 1, name: 'Charlie' },
    { id: 4, parentId: 2, name: 'David' },
    { id: 5, parentId: 2, name: 'Eve' },
    { id: 6, parentId: 3, name: 'Frank' }
  ]

  window.chart = new OrgChart().container('#chart');
  window.chart.nodeContent(renderNodeContent)

  window.chart.data(data).render();
}

function renderNodeContent(d) {
  return `<div class="person">
            <h3>${d.data.name}</h3>
            <p>ID: ${d.data.id}</p>
          </div>`;
}
