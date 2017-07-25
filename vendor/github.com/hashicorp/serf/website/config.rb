set :base_url, "https://www.serf.io/"

activate :hashicorp do |h|
  h.name        = "serf"
  h.version     = "0.8.1"
  h.github_slug = "hashicorp/serf"
end

helpers do
  # Get the title for the page.
  #
  # @param [Middleman::Page] page
  #
  # @return [String]
  def title_for(page)
    if page && page.data.page_title
      return "#{page.data.page_title} - Serf by HashiCorp"
    end

     "Serf by HashiCorp"
   end

  # Get the description for the page
  #
  # @param [Middleman::Page] page
  #
  # @return [String]
  def description_for(page)
    return escape_html(page.data.description || "")
  end

  # This helps by setting the "active" class for sidebar nav elements
  # if the YAML frontmatter matches the expected value.
  def sidebar_current(expected)
    current = current_page.data.sidebar_current
    if current.start_with?(expected)
      return " class=\"active\""
    else
      return ""
    end
  end
end
