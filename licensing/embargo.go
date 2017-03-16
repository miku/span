package licensing

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidEmbargo when embargo cannot be interpreted.
	ErrInvalidEmbargo = errors.New("invalid embargo")

	// Day is fixed number of hours.
	Day = 24 * time.Hour
	// Month is fixed number of hours.
	Month = 730 * time.Hour
	// Year is fixed number of hours.
	Year = 8760 * time.Hour

	// embargoPattern fixes allowed embargo strings (type, length, units).
	embargoPattern = regexp.MustCompile(`([P|R])([0-9]+)([Y|M|D])`)

	ErrBeforeMovingWall = errors.New("before moving wall")
	ErrAfterMovingWall  = errors.New("after moving wall")
)

// Embargo holds moving wall information.
//
// http://www.uksg.org/kbart/s5/guidelines/data_fields
//
// The embargo field reflects limitations on when resources become available
// online, generally as a result of contractual limitations established between the
// publisher and the content provider. Presenting this information to librarians
// (usually via link resolver owners) is vital to ensure that link resolvers do not
// generate links to content that is not yet available for users to access.
//
// One of the biggest problems facing members of this supply chain is that multiple
// kinds of embargoes exist—in some cases, coverage "to one year ago" means
// that data from 365 days ago becomes available today, while in other cases it
// means that the item is not available until the end of the current calendar year.
//
// Because of the complexities of embargoes, we recommend that the ISO 8601
// date syntax should be used. This is flexible enough to allow multiple types of
// embargoes to be described.
//
// The following method for specifying embargoes is derived from the ISO 8601
// "duration syntax" standard, making a few additional distinctions not covered in
// the standard. The embargo statement has three parts: type, length, and units.
// These three parts are written in that order in a single string with no spaces.
//
//   * Type: All embargoes involve a "moving wall", a point in time expressed relative to
//           the present (e.g., "12 months ago"). If access to the journal begins at the moving
//           wall, the embargo type is "R". If access ends at the moving wall, then the
//           embargo type is "P".
//   * Length: An integer expressing the length of the embargo
//   * Units: The units for the number in the "length" field: "D" for days, "M" for months,
//            and "Y" for years. For simplicity, "365D" will always be equivalent to one year,
//            and "30D" will always be equivalent to one month, even in leap years and months
//            that do not have 30 days.
//
// The "units" field also indicates the granularity of the embargo, that is, how
// frequently the moving wall "moves". For example, a newspaper database may
// have a subscription model that gives customers access to exactly one year of
// past content. Each day, a new issue is added, and the issue that was published
// exactly one year ago that day is removed from the customer’s access. In this
// case, the embargo statement would be "R365D", because the date of the earliest
// accessible issue changes each day.
//
// Another journal may have a model that gives access to all issues in the current
// year, starting in January. The following January, the customer loses access to all
// of the previous year’s issues at once, and will only be able to access issues
// published in the current year. In this case, we would say that the customer has
// access to one "calendar year" of content. The embargo statement would be
// "R1Y", because the date of the earliest issue changes once a year.
//
// Below are some common embargoes expressed according to this syntax:
//
//   * Access to all content, except the current calendar year: P1Y
//   * Access to all content in the previous and current calendar years: R2Y
//   * Access to all content from exactly 6 months ago to the present: R180D
//   * Access to all content, except the past 6 calendar months: P6M
//
// In the case where there is an embargo at both the beginning and end of a
// coverage range, then two embargo statements should be concatenated, the
// starting embargo coming first. The two statements should be separated by a
// semicolon. For example, "R10Y;P30D" describes an archive in which the past 10
// calendar years of content are available, except for the most current 30 days.
type Embargo string

// Duration converts embargo like P12M, P1M, R10Y into a time.Duration. This
// duration will non-negative. Time differences will have small shifts due to a
// month and a year being a fixed number of hours.
func (embargo Embargo) Duration() (dur time.Duration, err error) {
	e := strings.TrimSpace(string(embargo))
	if len(e) == 0 {
		return
	}

	var parts = embargoPattern.FindStringSubmatch(e)
	if len(parts) == 0 {
		return dur, ErrInvalidEmbargo
	}
	if len(parts) < 4 {
		return dur, ErrInvalidEmbargo
	}
	i, err := strconv.Atoi(parts[2])
	if err != nil {
		return dur, ErrInvalidEmbargo
	}

	switch parts[3] {
	case "D":
		return time.Duration(i) * Day, nil
	case "M":
		return time.Duration(i) * Month, nil
	case "Y":
		return time.Duration(i) * Year, nil
	default:
		return dur, ErrInvalidEmbargo
	}
}

// AccessBeginsAtWall returns true, if access begins at the moving wall.
func (embargo Embargo) AccessBeginsAtWall() bool {
	return strings.HasPrefix(strings.TrimSpace(string(embargo)), "R")
}

// AccessEndsAtWall returns true, if access end at the moving wall.
func (embargo Embargo) AccessEndsAtWall() bool {
	return strings.HasPrefix(strings.TrimSpace(string(embargo)), "P")
}

// Compatible returns true, if the given date is validated by the embargo relative to the current time.
func (embargo Embargo) Compatible(t time.Time) error {
	return embargo.CompatibleTo(t, time.Now())
}

// CompatibleTo returns true, if the given date in validated by this embargo relative to another date.
func (embargo Embargo) CompatibleTo(t time.Time, relative time.Time) error {
	dur, err := embargo.Duration()
	if err != nil {
		return err
	}
	wall := relative.Add(-dur)
	if embargo.AccessBeginsAtWall() && t.Before(wall) {
		return ErrBeforeMovingWall
	}
	if embargo.AccessEndsAtWall() && t.After(wall) {
		return ErrAfterMovingWall
	}
	return nil
}
