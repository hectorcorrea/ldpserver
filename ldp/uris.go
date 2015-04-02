package ldp

const (
	RdfTypeUri = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
)

const (
	LdpResourceUri       = "http://www.w3.org/ns/ldp#Resource"
	LdpRdfSourceUri      = "http://www.w3.org/ns/ldp#RDFSource"
	LdpNonRdfSourceUri   = "http://www.w3.org/ns/ldp#NonRDFSource"
	LdpContainerUri      = "http://www.w3.org/ns/ldp#Container"
	LdpBasicContainerUri = "http://www.w3.org/ns/ldp#BasicContainer"
	LdpContainsUri       = "http://www.w3.org/ns/ldp#Contains"
)

const (
	// HTTP header links
	LdpResourceLink       = LdpResourceUri + "; rel=\"type\""
	LdpNonRdfSourceLink   = LdpNonRdfSourceUri + "; rel=\"type\""
	LdpContainerLink      = LdpContainerUri + "; rel=\"type\""
	LdpBasicContainerLink = LdpBasicContainerUri + "; rel=\"type\""
)

const (
	DcTitleUri   = "http://purl.org/dc/terms/title"
	DcCreatedUri = "http://purl.org/dc/terms/created"
)
