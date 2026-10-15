package gbif

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"workshop/internal/common/httputil"
)

// OccurrencesSearchArgs represents the arguments required
// to search for occurrences.
type OccurrencesSearchArgs struct {
	PaginationOptions
	Country     string
	Latitude    float64
	Longitude   float64
	RadiusMiles float64
	DateFrom    time.Time
	DateTo      time.Time
	TaxonKeys   []int
}

// Occurrence retrieves a single occurrence by its GBIF key.
//
// See: https://techdocs.gbif.org/en/openapi/v1/occurrence#/Occurrence/getOccurrenceById
func (c Client) Occurrence(ctx context.Context, key int64) (Occurrence, error) {
	return c.http.Do[Occurrence](ctx, httputil.Request{
		Method: http.MethodGet,
		Path:   "/occurrence/" + strconv.FormatInt(key, 10),
	})
}

// Occurrences fetches GET /occurrence/search
//
// Obtains a collection of occurrences by search arguments.
// GBIF caps pages at 300 records and rejects offset + limit above 100,000.
//
// See: https://techdocs.gbif.org/en/openapi/v1/occurrence
func (c Client) Occurrences(ctx context.Context, args OccurrencesSearchArgs) (PaginatedResponse[Occurrence], error) {
	return c.http.Do[PaginatedResponse[Occurrence]](ctx, httputil.Request{
		Method: http.MethodGet,
		Path:   "/occurrence/search",
		Query:  args.query(),
	})
}

// query builds the URL query values for the search arguments,
// including pagination, taxa, country, lat/long and date range filters.
func (a OccurrencesSearchArgs) query() url.Values {
	q := a.PaginationOptions.query()
	seen := make(map[int]bool)
	for _, key := range a.TaxonKeys {
		if !seen[key] {
			q.Add("taxonKey", strconv.Itoa(key))
			seen[key] = true
		}
	}
	if len(seen) > 0 {
		// These IDs belong to the GBIF Backbone rather than Catalogue of Life.
		q.Set("checklistKey", "d7dddbf4-2cf0-4f39-9b2a-bb099caae36c")
	}

	if a.Country != "" {
		q.Set("country", a.Country)
	}

	switch {
	case a.RadiusMiles > 0:
		// geo_distance searches within a radius of a point, unlike
		// decimalLatitude/decimalLongitude which only match a point exactly.
		q.Set("geo_distance", strconv.FormatFloat(a.RadiusMiles, 'f', -1, 64)+"mi,"+
			strconv.FormatFloat(a.Latitude, 'f', -1, 64)+","+
			strconv.FormatFloat(a.Longitude, 'f', -1, 64))
	case a.Latitude != 0 || a.Longitude != 0:
		if a.Latitude != 0 {
			q.Set("decimalLatitude", strconv.FormatFloat(a.Latitude, 'f', -1, 64))
		}
		if a.Longitude != 0 {
			q.Set("decimalLongitude", strconv.FormatFloat(a.Longitude, 'f', -1, 64))
		}
	}

	// The GBIF API prefers a range of dates instead of from to,
	// so we need to convert it.
	var dates []string
	if !a.DateFrom.IsZero() {
		dates = append(dates, a.DateFrom.Format(dateLayout))
	}
	if !a.DateTo.IsZero() {
		dates = append(dates, a.DateTo.Format(dateLayout))
	}
	if len(dates) > 0 {
		q.Set("eventDate", strings.Join(dates, ","))
	}

	return q
}

// OccurrenceStatus describes the presence or absence of an occurrence
// at a location and time.
type OccurrenceStatus string

const (
	// OccurrenceStatusPresent means the species was present at the location and time.
	OccurrenceStatusPresent OccurrenceStatus = "PRESENT"

	// OccurrenceStatusAbsent means a survey concluded the species was absent.
	OccurrenceStatusAbsent OccurrenceStatus = "ABSENT"

	// OccurrenceStatusDetected means a survey method detected evidence of the species.
	OccurrenceStatusDetected OccurrenceStatus = "DETECTED"

	// OccurrenceStatusNotDetected means a survey looked but did not detect the species;
	// it does not prove that the species was absent.
	OccurrenceStatusNotDetected OccurrenceStatus = "NOT_DETECTED"
)

// Occurrence defines an individual, interpreted occurrence
// for a specific species.
type Occurrence struct {
	Key                    int64     `json:"key"`
	DatasetKey             string    `json:"datasetKey"`
	PublishingOrgKey       string    `json:"publishingOrgKey"`
	DatasetCategory        []string  `json:"datasetCategory,omitempty"`
	InstallationKey        string    `json:"installationKey"`
	HostingOrganizationKey string    `json:"hostingOrganizationKey"`
	PublishingCountry      string    `json:"publishingCountry"`
	Protocol               string    `json:"protocol"`
	LastCrawled            time.Time `json:"lastCrawled"`
	LastParsed             time.Time `json:"lastParsed"`
	CrawlId                int       `json:"crawlId"`
	Extensions             struct {
		HttpRsGbifOrgTerms10Multimedia []struct {
			// Preserve the publisher's raw date; use Media.Created for the normalized timestamp.
			HttpPurlOrgDcTermsCreated      string `json:"http://purl.org/dc/terms/created"`
			HttpPurlOrgDcTermsRightsHolder string `json:"http://purl.org/dc/terms/rightsHolder"`
			HttpPurlOrgDcTermsType         string `json:"http://purl.org/dc/terms/type"`
			HttpPurlOrgDcTermsIdentifier   string `json:"http://purl.org/dc/terms/identifier"`
			HttpPurlOrgDcTermsTitle        string `json:"http://purl.org/dc/terms/title,omitempty"`
			HttpPurlOrgDcTermsLicense      string `json:"http://purl.org/dc/terms/license,omitempty"`
			HttpPurlOrgDcTermsFormat       string `json:"http://purl.org/dc/terms/format"`
			HttpPurlOrgDcTermsCreator      string `json:"http://purl.org/dc/terms/creator,omitempty"`
		} `json:"http://rs.gbif.org/terms/1.0/Multimedia,omitempty"`
		HttpRsTdwgOrgAcTermsMultimedia []struct {
			HttpPurlOrgDcTermsRights                           string `json:"http://purl.org/dc/terms/rights"`
			HttpRsTdwgOrgAcTermsVariantLiteral                 string `json:"http://rs.tdwg.org/ac/terms/variantLiteral"`
			HttpNsAdobeComXap10RightsOwner                     string `json:"http://ns.adobe.com/xap/1.0/rights/Owner"`
			HttpRsTdwgOrgAcTermsCaption                        string `json:"http://rs.tdwg.org/ac/terms/caption,omitempty"`
			HttpPurlOrgDcElements11Creator                     string `json:"http://purl.org/dc/elements/1.1/creator"`
			HttpRsTdwgOrgAcTermsAssociatedObservationReference string `json:"http://rs.tdwg.org/ac/terms/associatedObservationReference"`
			HttpPurlOrgDcElements11Type                        string `json:"http://purl.org/dc/elements/1.1/type"`
			HttpPurlOrgDcTermsIdentifier                       string `json:"http://purl.org/dc/terms/identifier"`
			HttpPurlOrgDcTermsFormat                           string `json:"http://purl.org/dc/terms/format"`
			HttpRsTdwgOrgAcTermsAccessURI                      string `json:"http://rs.tdwg.org/ac/terms/accessURI"`
			HttpPurlOrgDcTermsDescription                      string `json:"http://purl.org/dc/terms/description,omitempty"`
			HttpRsTdwgOrgAcTermsPhysicalSetting                string `json:"http://rs.tdwg.org/ac/terms/physicalSetting,omitempty"`
			HttpNsAdobeComXap10Rating                          string `json:"http://ns.adobe.com/xap/1.0/Rating,omitempty"`
			HttpRsTdwgOrgAcTermsResourceCreationTechnique      string `json:"http://rs.tdwg.org/ac/terms/resourceCreationTechnique,omitempty"`
		} `json:"http://rs.tdwg.org/ac/terms/Multimedia,omitempty"`
	} `json:"extensions"`
	BasisOfRecord    string           `json:"basisOfRecord"`
	OccurrenceStatus OccurrenceStatus `json:"occurrenceStatus"`
	Classifications  struct {
		Ddf754FD1934Cc9B35199906754A03B struct {
			Usage struct {
				Key                  string `json:"key"`
				Name                 string `json:"name"`
				Rank                 string `json:"rank"`
				Code                 string `json:"code"`
				Authorship           string `json:"authorship"`
				GenericName          string `json:"genericName,omitempty"`
				SpecificEpithet      string `json:"specificEpithet,omitempty"`
				FormattedName        string `json:"formattedName"`
				InfraspecificEpithet string `json:"infraspecificEpithet,omitempty"`
				InfragenericEpithet  string `json:"infragenericEpithet,omitempty"`
			} `json:"usage"`
			AcceptedUsage struct {
				Key                  string `json:"key"`
				Name                 string `json:"name"`
				Rank                 string `json:"rank"`
				Code                 string `json:"code"`
				GenericName          string `json:"genericName,omitempty"`
				SpecificEpithet      string `json:"specificEpithet,omitempty"`
				InfraspecificEpithet string `json:"infraspecificEpithet,omitempty"`
				FormattedName        string `json:"formattedName"`
				Authorship           string `json:"authorship,omitempty"`
				InfragenericEpithet  string `json:"infragenericEpithet,omitempty"`
			} `json:"acceptedUsage"`
			TaxonomicStatus string `json:"taxonomicStatus"`
			Classification  []struct {
				Key  string `json:"key"`
				Name string `json:"name"`
				Rank string `json:"rank"`
			} `json:"classification"`
			Issues                  []string `json:"issues"`
			IucnRedListCategoryCode string   `json:"iucnRedListCategoryCode,omitempty"`
		} `json:"7ddf754f-d193-4cc9-b351-99906754a03b"`
		D7Dddbf42Cf04F399B2ABb099Caae36C struct {
			Usage struct {
				Key                  string `json:"key"`
				Name                 string `json:"name"`
				Rank                 string `json:"rank"`
				Code                 string `json:"code,omitempty"`
				GenericName          string `json:"genericName,omitempty"`
				SpecificEpithet      string `json:"specificEpithet,omitempty"`
				InfraspecificEpithet string `json:"infraspecificEpithet,omitempty"`
				FormattedName        string `json:"formattedName"`
				Authorship           string `json:"authorship,omitempty"`
			} `json:"usage"`
			AcceptedUsage struct {
				Key                  string `json:"key"`
				Name                 string `json:"name"`
				Rank                 string `json:"rank"`
				Code                 string `json:"code,omitempty"`
				GenericName          string `json:"genericName,omitempty"`
				SpecificEpithet      string `json:"specificEpithet,omitempty"`
				InfraspecificEpithet string `json:"infraspecificEpithet,omitempty"`
				FormattedName        string `json:"formattedName"`
				Authorship           string `json:"authorship,omitempty"`
			} `json:"acceptedUsage"`
			TaxonomicStatus string `json:"taxonomicStatus"`
			Classification  []struct {
				Key  string `json:"key"`
				Name string `json:"name"`
				Rank string `json:"rank"`
			} `json:"classification"`
			Issues                  []string `json:"issues"`
			IucnRedListCategoryCode string   `json:"iucnRedListCategoryCode,omitempty"`
		} `json:"d7dddbf4-2cf0-4f39-9b2a-bb099caae36c"`
	} `json:"classifications"`
	TaxonKey               int     `json:"taxonKey"`
	KingdomKey             int     `json:"kingdomKey"`
	PhylumKey              int     `json:"phylumKey"`
	ClassKey               int     `json:"classKey"`
	OrderKey               int     `json:"orderKey"`
	FamilyKey              int     `json:"familyKey"`
	GenusKey               int     `json:"genusKey"`
	SpeciesKey             int     `json:"speciesKey,omitempty"`
	AcceptedTaxonKey       int     `json:"acceptedTaxonKey"`
	ScientificName         string  `json:"scientificName"`
	AcceptedScientificName string  `json:"acceptedScientificName"`
	Kingdom                string  `json:"kingdom"`
	Phylum                 string  `json:"phylum"`
	Order                  string  `json:"order"`
	Family                 string  `json:"family"`
	Genus                  string  `json:"genus"`
	Species                string  `json:"species,omitempty"`
	GenericName            string  `json:"genericName,omitempty"`
	SpecificEpithet        string  `json:"specificEpithet,omitempty"`
	InfraspecificEpithet   string  `json:"infraspecificEpithet,omitempty"`
	TaxonRank              string  `json:"taxonRank"`
	TaxonomicStatus        string  `json:"taxonomicStatus"`
	DecimalLatitude        float64 `json:"decimalLatitude,omitempty"`
	DecimalLongitude       float64 `json:"decimalLongitude,omitempty"`
	Continent              string  `json:"continent"`
	Gadm                   struct {
		Level0 struct {
			Gid  string `json:"gid"`
			Name string `json:"name"`
		} `json:"level0"`
		Level1 struct {
			Gid  string `json:"gid"`
			Name string `json:"name"`
		} `json:"level1"`
		Level2 struct {
			Gid  string `json:"gid"`
			Name string `json:"name"`
		} `json:"level2"`
		Level3 struct {
			Gid  string `json:"gid"`
			Name string `json:"name"`
		} `json:"level3"`
	} `json:"gadm"`
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day,omitempty"`
	// EventDate uses the start of partial dates and intervals. Missing times and
	// timezones use midnight and UTC respectively; missing dates remain zero.
	EventDate       time.Time `json:"eventDate"`
	StartDayOfYear  int       `json:"startDayOfYear,omitempty"`
	EndDayOfYear    int       `json:"endDayOfYear,omitempty"`
	Issues          []string  `json:"issues"`
	LastInterpreted time.Time `json:"lastInterpreted"`
	License         string    `json:"license"`
	IsSequenced     bool      `json:"isSequenced"`
	Identifiers     []struct {
		Identifier string `json:"identifier"`
	} `json:"identifiers"`
	Media []struct {
		Type         string    `json:"type"`
		Format       string    `json:"format"`
		Title        string    `json:"title,omitempty"`
		Created      time.Time `json:"created"`
		License      string    `json:"license,omitempty"`
		RightsHolder string    `json:"rightsHolder"`
		Identifier   string    `json:"identifier"`
		Creator      string    `json:"creator,omitempty"`
		Description  string    `json:"description,omitempty"`
	} `json:"media"`
	Facts                                  []any     `json:"facts"`
	Relations                              []any     `json:"relations"`
	IsInCluster                            bool      `json:"isInCluster"`
	RecordedBy                             string    `json:"recordedBy,omitempty"`
	DnaSequenceID                          []any     `json:"dnaSequenceID"`
	NucleotideSequence                     []any     `json:"nucleotideSequence"`
	GeodeticDatum                          string    `json:"geodeticDatum,omitempty"`
	Class                                  string    `json:"class"`
	CountryCode                            string    `json:"countryCode"`
	RecordedByIDs                          []any     `json:"recordedByIDs"`
	IdentifiedByIDs                        []any     `json:"identifiedByIDs"`
	GbifRegion                             string    `json:"gbifRegion"`
	Country                                string    `json:"country"`
	PublishedByGbifRegion                  string    `json:"publishedByGbifRegion"`
	HttpUnknownOrgDownload                 string    `json:"http://unknown.org/download,omitempty"`
	DatasetTitle                           string    `json:"datasetTitle,omitempty"`
	CatalogNumber                          string    `json:"catalogNumber,omitempty"`
	HttpUnknownOrgCrawlAttempt             string    `json:"http://unknown.org/crawl_attempt"`
	HttpUnknownOrgOrphanEndpoint           string    `json:"http://unknown.org/orphanEndpoint,omitempty"`
	InstitutionCode                        string    `json:"institutionCode,omitempty"`
	Locality                               string    `json:"locality"`
	HttpUnknownOrgStatus                   string    `json:"http://unknown.org/status,omitempty"`
	HttpUnknownOrgGriddedDataset           string    `json:"http://unknown.org/griddedDataset,omitempty"`
	GbifID                                 string    `json:"gbifID"`
	CollectionCode                         string    `json:"collectionCode,omitempty"`
	NetworkKeys                            []string  `json:"networkKeys,omitempty"`
	Sex                                    string    `json:"sex,omitempty"`
	LifeStage                              string    `json:"lifeStage,omitempty"`
	ScientificNameAuthorship               string    `json:"scientificNameAuthorship,omitempty"`
	IucnRedListCategory                    string    `json:"iucnRedListCategory,omitempty"`
	HigherGeography                        string    `json:"higherGeography,omitempty"`
	TypeStatus                             string    `json:"typeStatus,omitempty"`
	Modified                               time.Time `json:"modified"`
	InstitutionKey                         string    `json:"institutionKey,omitempty"`
	OtherCatalogNumbers                    string    `json:"otherCatalogNumbers,omitempty"`
	HttpUnknownOrgConceptualSchema         string    `json:"http://unknown.org/conceptualSchema,omitempty"`
	Identifier                             string    `json:"identifier,omitempty"`
	DynamicProperties                      string    `json:"dynamicProperties,omitempty"`
	HttpUnknownOrgMaxSearchResponseRecords string    `json:"http://unknown.org/maxSearchResponseRecords,omitempty"`
	OccurrenceID                           string    `json:"occurrenceID,omitempty"`
	HttpUnknownOrgCode                     string    `json:"http://unknown.org/code,omitempty"`
	IndividualCount                        *int      `json:"individualCount,omitempty"`
	StateProvince                          string    `json:"stateProvince,omitempty"`
	HigherClassification                   string    `json:"higherClassification,omitempty"`
	CoordinateUncertaintyInMeters          *float64  `json:"coordinateUncertaintyInMeters,omitempty"`
	References                             string    `json:"references,omitempty"`
	IdentifiedBy                           string    `json:"identifiedBy,omitempty"`
	ScientificNameID                       string    `json:"scientificNameID,omitempty"`
	FieldNumber                            string    `json:"fieldNumber,omitempty"`
	Language                               string    `json:"language,omitempty"`
	TaxonID                                string    `json:"taxonID,omitempty"`
	LocationAccordingTo                    string    `json:"locationAccordingTo,omitempty"`
	VernacularName                         string    `json:"vernacularName,omitempty"`
	LocationID                             string    `json:"locationID,omitempty"`
	BibliographicCitation                  string    `json:"bibliographicCitation,omitempty"`
	Preparations                           string    `json:"preparations,omitempty"`
	RightsHolder                           string    `json:"rightsHolder,omitempty"`
	RecordNumber                           string    `json:"recordNumber,omitempty"`
	Municipality                           string    `json:"municipality,omitempty"`
	OwnerInstitutionCode                   string    `json:"ownerInstitutionCode,omitempty"`
	OccurrenceRemarks                      string    `json:"occurrenceRemarks,omitempty"`
	CollectionID                           string    `json:"collectionID,omitempty"`
	NomenclaturalCode                      string    `json:"nomenclaturalCode,omitempty"`
	VerbatimEventDate                      string    `json:"verbatimEventDate,omitempty"`
	FieldNotes                             string    `json:"fieldNotes,omitempty"`
	EventTime                              string    `json:"eventTime,omitempty"`
	VerbatimElevation                      string    `json:"verbatimElevation,omitempty"`
	Behavior                               string    `json:"behavior,omitempty"`
	EstablishmentMeans                     string    `json:"establishmentMeans,omitempty"`
	DateIdentified                         string    `json:"dateIdentified,omitempty"`
	Elevation                              *float64  `json:"elevation,omitempty"`
	OrganismQuantity                       float64   `json:"organismQuantity,omitempty"`
	DatasetName                            string    `json:"datasetName,omitempty"`
	EventID                                string    `json:"eventID,omitempty"`
	FootprintWKT                           string    `json:"footprintWKT,omitempty"`
	County                                 string    `json:"county,omitempty"`
	VerbatimIdentification                 string    `json:"verbatimIdentification,omitempty"`
	FootprintSRS                           string    `json:"footprintSRS,omitempty"`
	TaxonConceptID                         string    `json:"taxonConceptID,omitempty"`
	HttpUnknownOrgSamplingEvent            string    `json:"http://unknown.org/samplingEvent,omitempty"`
	DatasetID                              string    `json:"datasetID,omitempty"`
	SamplingProtocol                       string    `json:"samplingProtocol,omitempty"`
	IdentificationVerificationStatus       string    `json:"identificationVerificationStatus,omitempty"`
	NomenclaturalStatus                    string    `json:"nomenclaturalStatus,omitempty"`
	VerbatimSRS                            string    `json:"verbatimSRS,omitempty"`
	VerbatimLocality                       string    `json:"verbatimLocality,omitempty"`
	ParentNameUsageID                      string    `json:"parentNameUsageID,omitempty"`
	HttpUnknownOrgOmitFromScheduledCrawl   string    `json:"http://unknown.org/omitFromScheduledCrawlomitempty"`
	AcceptedNameUsageID                    string    `json:"acceptedNameUsageID,omitempty"`
	AccessRights                           string    `json:"accessRights,omitempty"`
	OrganismID                             string    `json:"organismID,omitempty"`
	GeoreferenceVerificationStatus         string    `json:"georeferenceVerificationStatus,omitempty"`
	EventRemarks                           string    `json:"eventRemarks,omitempty"`
}

// UnmarshalJSON accepts GBIF's ISO 8601 event dates, including partial dates
// and intervals, while exposing EventDate as a time.Time to consumers.
func (o *Occurrence) UnmarshalJSON(data []byte) error {
	type occurrence Occurrence
	decoded := occurrence(*o)
	wire := struct {
		*occurrence
		EventDate *string `json:"eventDate"`
	}{occurrence: &decoded}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.EventDate != nil {
		date, err := parseEventDate(*wire.EventDate)
		if err != nil {
			return err
		}
		decoded.EventDate = date
	}
	*o = Occurrence(decoded)
	return nil
}

func parseEventDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) > 2 {
		return time.Time{}, fmt.Errorf("gbif: invalid eventDate %q", value)
	}
	var start time.Time
	for i, part := range parts {
		var date time.Time
		var err error
		for _, layout := range []string{
			time.RFC3339Nano,
			"2006-01-02T15:04Z07:00",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			dateLayout,
			"2006-01",
			"2006",
		} {
			date, err = time.Parse(layout, part)
			if err == nil {
				break
			}
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("gbif: invalid eventDate %q: %w", value, err)
		}
		if i == 0 {
			start = date
		} else if date.Before(start) {
			return time.Time{}, fmt.Errorf("gbif: eventDate interval ends before it starts: %q", value)
		}
	}
	return start, nil
}
