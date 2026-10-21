import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { ClientToFrontend } from '../gen/proto/orgviz/clientapi/v1/clientapi_pb'

const transport = createConnectTransport({
  baseUrl: '/api/'
})

export const client = createClient(ClientToFrontend, transport)
