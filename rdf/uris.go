package rdf

const (
	RdfTypeUri = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
)

const (
	LdpResourceUri        = "http://www.w3.org/ns/ldp#Resource"
	LdpRdfSourceUri       = "http://www.w3.org/ns/ldp#RDFSource"
	LdpNonRdfSourceUri    = "http://www.w3.org/ns/ldp#NonRDFSource"
	LdpContainerUri       = "http://www.w3.org/ns/ldp#Container"
	LdpBasicContainerUri  = "http://www.w3.org/ns/ldp#BasicContainer"
	LdpDirectContainerUri = "http://www.w3.org/ns/ldp#DirectContainer"
	LdpContainsUri        = "http://www.w3.org/ns/ldp#Contains"
	LdpMembershipResource = "http://www.w3.org/ns/ldp#membershipResource"
	LdpHasMemberRelation  = "http://www.w3.org/ns/ldp#hasMemberRelation"
)

const (
	// HTTP header links
	LdpResourceLink        = "<" + LdpResourceUri + ">; rel=\"type\""
	LdpNonRdfSourceLink    = "<" + LdpNonRdfSourceUri + ">; rel=\"type\""
	LdpContainerLink       = "<" + LdpContainerUri + ">; rel=\"type\""
	LdpBasicContainerLink  = "<" + LdpBasicContainerUri + ">; rel=\"type\""
	LdpDirectContainerLink = "<" + LdpDirectContainerUri + ">; rel=\"type\""
)

const (
	DcTitleUri   = "http://purl.org/dc/terms/title"
	DcCreatedUri = "http://purl.org/dc/terms/created"
)

const (
	NTripleContentType = "text/ntriple"
	TurtleContentType  = "text/turtle"
)
