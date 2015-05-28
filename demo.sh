#!/bin/sh

# This script shows how to create a blog with one entry 
# and two comments for the entry. 
#
# The resulting URLs will look more or less as follows:
#     /blog1
#     /blog1/entry2
#     /blog1/entry2/content3
#     /blog1/entry2/comments4
#     /blog1/entry2/comments4/comment5
#     /blog1/entry2/comments4/comment6
#
#
# Make sure the LDP server is running on localhost:9001
#

# Create a blog
BLOG_URI="$(curl -X POST --header 'Slug: blog' localhost:9001)"

# Add an entry to the blog
ENTRY_URI="$(curl -X POST --header 'Slug: entry' -d '<> <dc:title> "blog one title" .' ${BLOG_URI})"

# Add the content for the blog (non-RDF)
CONTENT_URI="$(curl -X POST --header 'Slug: content' --header 'Link: http://www.w3.org/ns/ldp#NonRDFSource; rel=\"type\"' --data 'content of the blog entry' ${BLOG_URI})"

# Add a direct container for comments (and bound to the entry)
COMMENTS_URI="$(curl -X POST --header 'Slug: comments' ${ENTRY_URI})"
TRIPLE1="<> <http://www.w3.org/ns/ldp#hasMemberRelation> <hasComment> ."
TRIPLE2="<> <http://www.w3.org/ns/ldp#membershipResource> <${ENTRY_URI}> ."
curl -X PATCH -d "${TRIPLE1}" ${COMMENTS_URI}
curl -X PATCH -d "${TRIPLE2}" ${COMMENTS_URI}

# # Add a couple of comments to the direct container 
COMMENT1_URI="$(curl -X POST --header 'Slug: comment' -d $'<> <dc:description> "this is a comment" .' ${COMMENTS_URI})"
COMMENT2_URI="$(curl -X POST --header "Slug: comment" -d $'<> <dc:description> "this is another comment" .' ${COMMENTS_URI})"

echo "** The following URIs were created:"
echo "  BLOG_URI     = ${BLOG_URI}"
echo "  ENTRY_URI    = ${ENTRY_URI}"
echo "  CONTENT_URI  = ${CONTENT_URI}"
echo "  COMMENTS_URI = ${COMMENTS_URI}"
echo "  COMMENT1_URI = ${COMMENT1_URI}"
echo "  COMMENT2_URI = ${COMMENT2_URI}"

echo "** Direct container:"
curl ${COMMENTS_URI}

echo "** Blog entry:"
curl ${ENTRY_URI}