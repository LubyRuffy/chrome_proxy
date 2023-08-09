package render_dom

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/chrome_proxy/chrome_action"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"log"
)

// RenderDom 生成单个url的 dom html
func RenderDom(options *models.ChromeParam) (*models.RenderDomOutput, error) {
	log.Println("RenderDom of url:", options.URL)

	var html string
	var actions []chromedp.Action

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	}))

	var title string
	actions = append(actions, chromedp.Title(&title))
	var location string
	actions = append(actions, chromedp.Location(&location))

	err := chrome_action.ChromeActions(options.ChromeActionInput, func(s string, i ...interface{}) {

	}, options.Timeout, nil, actions...)

	if err != nil {
		return nil, fmt.Errorf("RenderDom failed(%w): %s", err, options.URL)
	}

	return &models.RenderDomOutput{
		Html:     html,
		Title:    title,
		Location: location,
	}, nil
}
