package events

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// DateFilter represents a valid filter to use for filtering
// events by their date during a list.
type DateFilter string

const (
	DateFilterGT  DateFilter = "gt"
	DateFilterGTE DateFilter = "gte"
	DateFilterLT  DateFilter = "lt"
	DateFilterLTE DateFilter = "lte"
)

// DateQuery represents a date field to be used for listing events.
// If no filter is specified, the query will act as if "equal" is used.
type DateQuery struct {
	Date   time.Time
	Filter DateFilter
}

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToEventListQuery() (string, error)
}

// ListOpts allows the filtering of paginated collections through the API.
// Filtering is achieved by passing in filter value. Page and PerPage are used
// for pagination.
type ListOpts struct {
	ObserverType string `q:"observer_type"`
	TargetID     string `q:"target_id"`
	TargetType   string `q:"target_type"`
	InitiatorID  string `q:"initiator_id"`

	// Not available for sort
	InitiatorType string `q:"initiator_type"`
	InitiatorName string `q:"initiator_name"`

	Action    string `q:"action"`
	Outcome   string `q:"outcome"`
	Time      []DateQuery
	DomainID  string `q:"domain_id"`
	ProjectID string `q:"project_id"`

	// Sort will sort the results in the requested order.
	Sort string `q:"sort"`

	Limit  int `q:"limit"`
	Offset int `q:"offset"`
}

// ToEventListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToEventListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	params := q.Query()

	if len(opts.Time) > 0 {
		var t []string

		for _, dt := range opts.Time {
			tmp := dt.Date.Format(time.RFC3339)
			if dt.Filter == "" {
				// combine gte and lte, when there is no filter
				t = append(t, fmt.Sprintf("%s:%s,%s:%s", DateFilterGTE, tmp, DateFilterLTE, tmp))
			} else {
				t = append(t, fmt.Sprintf("%s:%s", dt.Filter, tmp))
			}
		}

		params.Add("time", strings.Join(t, ","))
	}

	q = &url.URL{RawQuery: params.Encode()}

	return q.String(), nil
}

// List retrieves a list of Events.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := listURL(client)
	if opts != nil {
		query, err := opts.ToEventListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return EventPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get retrieves a specific event based on its unique ID.
func Get(c *gophercloud.ServiceClient, id string) (r GetResult) {
	_, r.Err = c.Get(getURL(c, id), &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	return
}
