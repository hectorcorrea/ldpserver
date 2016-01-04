#!/bin/sh

# This script shows how to create a blog with one entry
# and two comments for the entry.
#
# The resulting URLs will look more or less as follows:
#     /node1
#     /node1/entry
#     /node1/entry/content
#     /node1/entry/comments
#     /node1/entry/comments/comment1
#     /node1/entry/comments/comment2
#
# Make sure the LDP server is running on localhost:9001
#

# Create a blog
BLOG_URI="$(curl -X POST localhost:9001)"

# Add an entry to the blog
ENTRY_URI="$(curl -X POST --header "Content-Type: text/turtle" --header 'Slug: entry' -d '<> dc:title "blog one title" .' ${BLOG_URI})"

# Add the content for the blog (non-RDF)
CONTENT_URI="$(curl -X POST --header "Content-Type: text/plain" --header 'Slug: content' -d 'content of the blog entry' ${BLOG_URI})"

# Create a direct container for comments...
COMMENTS_URI="$(curl -X POST --header 'Slug: comments' ${ENTRY_URI})"

# ...and bind it to the entry
TRIPLE1="<> <http://www.w3.org/ns/ldp#hasMemberRelation> hasComment ."
TRIPLE2="<> <http://www.w3.org/ns/ldp#membershipResource> <${ENTRY_URI}> ."
curl -X PATCH --header "Content-Type: text/turtle" -d "${TRIPLE1}" ${COMMENTS_URI}
curl -X PATCH --header "Content-Type: text/turtle" -d "${TRIPLE2}" ${COMMENTS_URI}

# # Add a couple of comments to the direct container
COMMENT1_URI="$(curl -X POST --header "Content-Type: text/turtle" --header 'Slug: comment1' -d $'<> dc:description "this is a comment" .' ${COMMENTS_URI})"
COMMENT2_URI="$(curl -X POST --header "Content-Type: text/turtle" --header "Slug: comment2" -d $'<> dc:description "this is another comment" .' ${COMMENTS_URI})"

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