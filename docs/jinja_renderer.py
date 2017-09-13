import json

configuration = json.load(file('_static/configuration.json'))

html_context = {
    'config_options': configuration
}

def rstjinja(app, docname, source):
    """
    Render our pages as a jinja template for fancy templating goodness.
    """
    # Make sure we're outputting HTML
    if app.builder.format != 'html':
        return
    src = source[0]
    rendered = app.builder.templates.render_string(
        src, html_context
    )
    source[0] = rendered

def setup(app):
    app.connect("source-read", rstjinja)
