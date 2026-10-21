import * as d3 from 'd3'
import { OrgChart } from 'd3-org-chart'

export function renderNodeContent (d) {
  return `<div class="person">
            <img src="${d.data.avatarUrl}" alt="${d.data.fullName}" class="avatar" />
            <div class="person-info">
            <h3>${d.data.fullName}</h3>
            <p>${d.data.jobTitle}</p>
            </div>
          </div>`
}

export function createOrgChart (container, people, isCompact) {
  const chart = new OrgChart()
    .nodeHeight(() => 85)
    .container(container)
    .childrenMargin(() => 50)
    .compactMarginBetween(() => 35)
    .compactMarginPair(() => 30)
    .neighbourMargin(() => 20)
    .compact(isCompact)
    .linkUpdate(function () {
      return d3.select(this).attr('stroke', '#666')
    })
    .nodeContent(renderNodeContent)

  chart.data(people).render()
  return chart
}
