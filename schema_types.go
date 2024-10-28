package main

// TermMeta represents the termmeta table.
type TermMeta struct {
    MetaID    int64  `json:"meta_id"`
    TermID    int64  `json:"term_id"`
    MetaKey   *string `json:"meta_key,omitempty"`
    MetaValue *string `json:"meta_value,omitempty"`
}

// Terms represents the terms table.
type Terms struct {
    TermID    int64  `json:"term_id"`
    Name      string `json:"name"`
    Slug      string `json:"slug"`
    TermGroup int64  `json:"term_group"`
}

// TermTaxonomy represents the term_taxonomy table.
type TermTaxonomy struct {
    TermTaxonomyID int64  `json:"term_taxonomy_id"`
    TermID         int64  `json:"term_id"`
    Taxonomy       string `json:"taxonomy"`
    Description    string `json:"description"`
    Parent         int64  `json:"parent"`
    Count          int64  `json:"count"`
}

// TermRelationships represents the term_relationships table.
type TermRelationships struct {
    ObjectID        int64 `json:"object_id"`
    TermTaxonomyID  int64 `json:"term_taxonomy_id"`
    TermOrder       int   `json:"term_order"`
}

// CommentMeta represents the commentmeta table.
type CommentMeta struct {
    MetaID     int64   `json:"meta_id"`
    CommentID  int64   `json:"comment_id"`
    MetaKey    *string `json:"meta_key,omitempty"`
    MetaValue  *string `json:"meta_value,omitempty"`
}

// Comments represents the comments table.
type Comments struct {
    CommentID           int64   `json:"comment_id"`
    CommentPostID       int64   `json:"comment_post_id"`
    CommentAuthor       string  `json:"comment_author"`
    CommentAuthorEmail  string  `json:"comment_author_email"`
    CommentAuthorURL    string  `json:"comment_author_url"`
    CommentAuthorIP     string  `json:"comment_author_ip"`
    CommentDate         string  `json:"comment_date"`
    CommentDateGMT      string  `json:"comment_date_gmt"`
    CommentContent      string  `json:"comment_content"`
    CommentKarma        int     `json:"comment_karma"`
    CommentApproved     string  `json:"comment_approved"`
    CommentAgent        string  `json:"comment_agent"`
    CommentType         string  `json:"comment_type"`
    CommentParent       int64   `json:"comment_parent"`
    UserID              int64   `json:"user_id"`
}

// Links represents the links table.
type Links struct {
    LinkID       int64   `json:"link_id"`
    LinkURL      string  `json:"link_url"`
    LinkName     string  `json:"link_name"`
    LinkImage    string  `json:"link_image"`
    LinkTarget   string  `json:"link_target"`
    LinkDescription string `json:"link_description"`
    LinkVisible  string  `json:"link_visible"`
    LinkOwner    int64   `json:"link_owner"`
    LinkRating   int     `json:"link_rating"`
    LinkUpdated  string  `json:"link_updated"`
    LinkRel      string  `json:"link_rel"`
    LinkNotes    *string `json:"link_notes,omitempty"`
    LinkRSS      string  `json:"link_rss"`
}

// Options represents the options table.
type Options struct {
    OptionID    int64   `json:"option_id"`
    OptionName  string  `json:"option_name"`
    OptionValue string  `json:"option_value"`
    Autoload    string  `json:"autoload"`
}

// PostMeta represents the postmeta table.
type PostMeta struct {
    MetaID    int64   `json:"meta_id"`
    PostID    int64   `json:"post_id"`
    MetaKey   *string `json:"meta_key,omitempty"`
    MetaValue *string `json:"meta_value,omitempty"`
}

// Posts represents the posts table.
type Posts struct {
    ID                   int64   `json:"id"`
    PostAuthor           int64   `json:"post_author"`
    PostDate             string  `json:"post_date"`
    PostDateGMT          string  `json:"post_date_gmt"`
    PostContent          string  `json:"post_content"`
    PostTitle            string  `json:"post_title"`
    PostExcerpt          string  `json:"post_excerpt"`
    PostStatus           string  `json:"post_status"`
    CommentStatus        string  `json:"comment_status"`
    PingStatus           string  `json:"ping_status"`
    PostPassword         string  `json:"post_password"`
    PostName             string  `json:"post_name"`
    ToPing               string  `json:"to_ping"`
    Pinged               string  `json:"pinged"`
    PostModified         string  `json:"post_modified"`
    PostModifiedGMT      string  `json:"post_modified_gmt"`
    PostContentFiltered  string  `json:"post_content_filtered"`
    PostParent           int64   `json:"post_parent"`
    GUID                 string  `json:"guid"`
    MenuOrder            int     `json:"menu_order"`
    PostType             string  `json:"post_type"`
    PostMimeType         string  `json:"post_mime_type"`
    CommentCount         int64   `json:"comment_count"`
}

// Users represents the users table.
type Users2 struct {
    ID                 int64   `json:"id"`
    UserLogin          string  `json:"user_login"`
    UserPass           string  `json:"user_pass"`
    UserNicename       string  `json:"user_nicename"`
    UserEmail          string  `json:"user_email"`
    UserURL            string  `json:"user_url"`
    UserRegistered     string  `json:"user_registered"`
    UserActivationKey  string  `json:"user_activation_key"`
    UserStatus         int     `json:"user_status"`
    DisplayName        string  `json:"display_name"`
}


// UserMeta represents the usermeta table.
type UserMeta struct {
    UMetaID   int64   `json:"umeta_id"`
    UserID    int64   `json:"user_id"`
    MetaKey   *string `json:"meta_key,omitempty"`
    MetaValue *string `json:"meta_value,omitempty"`
}

