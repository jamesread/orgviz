package exportchart

import "github.com/jamesread/orgviz/internal/orgs"

const (
	cardWidth  = 250
	hGap       = 20
	vGap       = 50
	pairGap    = 30
	pairRowGap = 35
	padY       = 16
	padX       = 12
	avatarSize = 64
	avatarGap  = 12
	charWidth  = 7
	nameLine   = 18
	titleLine  = 16
	teamLine   = 14
	chartPad   = 16
)

type placed struct {
	person     *orgs.Person
	x, y, w, h int
}

type link struct {
	x1, y1, x2, y2 int
}

type chartLayout struct {
	cards []placed
	links []link
	w, h  int
}

type options struct {
	compact  bool
	features orgs.ResolvedFeatures
}

func layoutChart(chart *orgs.File, style orgs.Style) chartLayout {
	root := tree(chart.People)
	if root == nil {
		return chartLayout{}
	}

	opt := options{compact: style.Features.Resolve().Compact, features: style.Features.Resolve()}
	laid := place(root, chartPad, chartPad, opt)
	laid.w += chartPad
	laid.h += chartPad

	return laid
}

func place(node *personNode, left, top int, opt options) chartLayout {
	if len(node.children) == 0 {
		return leafLayout(node.person, left, top, opt)
	}

	return placeChildrenOf(node, left, top, opt)
}

func placeChildrenOf(node *personNode, left, top int, opt options) chartLayout {
	if opt.compact && allLeaves(node) && len(node.children) >= 2 {
		return placeCompact(node, left, top, opt)
	}

	return placeRow(node, left, top, opt)
}

func leafLayout(person *orgs.Person, left, top int, opt options) chartLayout {
	w, h := cardSize(person, opt)
	return chartLayout{
		cards: []placed{{person: person, x: left, y: top, w: w, h: h}},
		w:     left + w,
		h:     top + h,
	}
}

func cardSize(person *orgs.Person, opt options) (int, int) {
	return cardWidth, cardHeight(person, opt)
}

func cardHeight(person *orgs.Person, opt options) int {
	maxChars := textChars(opt.features.Avatars)
	body := nameLine * max(1, len(wrap(person.Name, maxChars)))
	body += titleLine * featureLines(opt.features.JobTitles, person.Title, maxChars)
	body += teamLine * featureLines(opt.features.Team, person.Team, maxChars)
	if opt.features.Avatars {
		body = max(body, avatarSize)
	}

	return body + padY*2
}

func featureLines(on bool, text string, maxChars int) int {
	if !on {
		return 0
	}

	return len(wrap(text, maxChars))
}

func textChars(showAvatar bool) int {
	width := cardWidth - padX*2
	if showAvatar {
		width -= avatarSize + avatarGap
	}

	return max(8, width/charWidth)
}

func placeRow(node *personNode, left, top int, opt options) chartLayout {
	cw, ch := cardSize(node.person, opt)
	kids, kidsW := placeChildren(node.children, left, top+ch+vGap, opt)
	width := max(cw, kidsW)
	parentX := left + (width-cw)/2
	laid := chartLayout{
		cards: []placed{{person: node.person, x: parentX, y: top, w: cw, h: ch}},
		w:     left + width,
		h:     top + ch,
	}

	return mergeChildren(laid, kids, parentX+cw/2, top+ch)
}

func placeChildren(children []*personNode, left, top int, opt options) (chartLayout, int) {
	laid := chartLayout{h: top}
	x := left
	for i, child := range children {
		if i > 0 {
			x += hGap
		}
		part := place(child, x, top, opt)
		laid = appendLayout(laid, part)
		x = part.w
	}

	return laid, x - left
}

func placeCompact(node *personNode, left, top int, opt options) chartLayout {
	cw, ch := cardSize(node.person, opt)
	pairW := compactWidth(len(node.children), cw)
	parentX := left + (pairW-cw)/2
	laid := chartLayout{
		cards: []placed{{person: node.person, x: parentX, y: top, w: cw, h: ch}},
		w:     left + pairW,
		h:     top + ch,
	}
	kids := placeCompactKids(node.children, left, top+ch+vGap, cw, opt)

	return mergeChildren(laid, kids, parentX+cw/2, top+ch)
}

func compactWidth(n, cardW int) int {
	if n < 2 {
		return cardW
	}

	return cardW*2 + pairGap
}

func placeCompactKids(children []*personNode, left, top, cardW int, opt options) chartLayout {
	laid := chartLayout{w: left + compactWidth(len(children), cardW)}
	for i, child := range children {
		cw, ch := cardSize(child.person, opt)
		x := left + (i%2)*(cardW+pairGap)
		y := top + (i/2)*(ch+pairRowGap)
		laid.cards = append(laid.cards, placed{person: child.person, x: x, y: y, w: cw, h: ch})
		laid.h = max(laid.h, y+ch)
		if i >= 2 {
			above := laid.cards[i-2]
			laid.links = append(laid.links, link{x1: x + cw/2, y1: above.y + above.h, x2: x + cw/2, y2: y})
		}
	}

	return laid
}

func mergeChildren(parent, kids chartLayout, parentCX, parentBottom int) chartLayout {
	parent.cards = append(parent.cards, kids.cards...)
	parent.links = append(parent.links, kids.links...)
	parent.links = append(parent.links, childLinks(parentCX, parentBottom, kids.cards)...)
	parent.w = max(parent.w, kids.w)
	parent.h = max(parent.h, kids.h)

	return parent
}

func childLinks(parentCX, parentBottom int, cards []placed) []link {
	if len(cards) == 0 {
		return nil
	}

	hubY := parentBottom + vGap/2
	links := []link{{x1: parentCX, y1: parentBottom, x2: parentCX, y2: hubY}}
	for _, card := range firstRow(cards, cards[0].y) {
		cx := card.x + card.w/2
		links = append(links, link{x1: parentCX, y1: hubY, x2: cx, y2: hubY})
		links = append(links, link{x1: cx, y1: hubY, x2: cx, y2: card.y})
	}

	return links
}

func firstRow(cards []placed, y int) []placed {
	row := []placed{}
	for _, card := range cards {
		if card.y == y {
			row = append(row, card)
		}
	}

	return row
}

func appendLayout(base, extra chartLayout) chartLayout {
	base.cards = append(base.cards, extra.cards...)
	base.links = append(base.links, extra.links...)
	base.w = max(base.w, extra.w)
	base.h = max(base.h, extra.h)

	return base
}
