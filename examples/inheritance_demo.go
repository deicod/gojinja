package examples

import (
	"fmt"
	"log"

	"github.com/deicod/gojinja/runtime"
)

func RunInheritanceDemo() {
	// Create a new environment
	env := runtime.NewEnvironment()

	// Create a map loader for demonstration
	templates := map[string]string{
		"base.html": `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    <header>
        <h1>{% block header %}Default Header{% endblock %}</h1>
    </header>

    <main>
        {% block content %}{% endblock %}
    </main>

    <footer>
        {% block footer %}
            <p>&copy; 2024 My Website. All rights reserved.</p>
        {% endblock %}
    </footer>
</body>
</html>`,

		"page.html": `{% extends "base.html" %}

{% block title %}{{ page_title }} - My Site{% endblock %}

{% block header %}Welcome to {{ site_name }}{% endblock %}

{% block content %}
    <h2>{{ page_title }}</h2>
    <p>{{ content }}</p>

    {% block sidebar %}
        <aside>
            <h3>Navigation</h3>
            <ul>
                <li><a href="/">Home</a></li>
                <li><a href="/about">About</a></li>
                <li><a href="/contact">Contact</a></li>
            </ul>
        </aside>
    {% endblock %}
{% endblock %}`,

		"about.html": `{% extends "page.html" %}

{% block title %}About Us - {{ site_name }}{% endblock %}

{% block content %}
    {{ super() }}

    <div class="about-content">
        <h3>Our Story</h3>
        <p>{{ about_text }}</p>

        {% block team %}
            <h4>Our Team</h4>
            <ul>
                {% for member in team_members %}
                    <li>{{ member.name }} - {{ member.role }}</li>
                {% endfor %}
            </ul>
        {% endblock %}
    </div>
{% endblock %}`,

		"contact.html": `{% extends "page.html" %}

{% block title %}Contact Us - {{ site_name }}{% endblock %}

{% block content %}
    <h2>Contact Information</h2>
    <form method="post">
        <div>
            <label for="name">Name:</label>
            <input type="text" id="name" name="name" required>
        </div>
        <div>
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div>
            <label for="message">Message:</label>
            <textarea id="message" name="message" required></textarea>
        </div>
        <button type="submit">Send</button>
    </form>
{% endblock %}

{% block sidebar %}
    <aside>
        <h3>Contact Details</h3>
        <p>
            <strong>Email:</strong> {{ contact_email }}<br>
            <strong>Phone:</strong> {{ contact_phone }}<br>
            <strong>Address:</strong> {{ contact_address }}
        </p>
    </aside>
    {{ super() }}
{% endblock %}`,
	}

	// Set up the map loader
	loader := runtime.NewMapLoader(templates)
	env.SetLoader(loader)

	// Test basic template inheritance
	fmt.Println("=== Basic Template Inheritance ===")
	testBasicInheritance(env)

	// Test nested inheritance with super()
	fmt.Println("\n=== Nested Inheritance with Super() ===")
	testNestedInheritance(env)

	// Test block overriding with super()
	fmt.Println("\n=== Block Overriding with Super() ===")
	testBlockOverriding(env)

	// Test template caching
	fmt.Println("\n=== Template Caching ===")
	testTemplateCaching(env)
}

func testBasicInheritance(env *runtime.Environment) {
	context := map[string]interface{}{
		"page_title": "Welcome",
		"site_name":  "My Awesome Site",
		"content":    "This is the main content of our page.",
	}

	tmpl, err := env.ParseFile("page.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	result, err := tmpl.ExecuteToString(context)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	fmt.Println(result)
}

func testNestedInheritance(env *runtime.Environment) {
	context := map[string]interface{}{
		"site_name":    "My Awesome Site",
		"page_title":   "About Us",
		"content":      "Learn more about our company.",
		"about_text":   "We are a company dedicated to excellence.",
		"team_members": []map[string]interface{}{
			{"name": "Alice Johnson", "role": "CEO"},
			{"name": "Bob Smith", "role": "CTO"},
			{"name": "Carol Davis", "role": "Designer"},
		},
	}

	tmpl, err := env.ParseFile("about.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	result, err := tmpl.ExecuteToString(context)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	fmt.Println(result)
}

func testBlockOverriding(env *runtime.Environment) {
	context := map[string]interface{}{
		"site_name":      "My Awesome Site",
		"page_title":     "Contact Us",
		"contact_email":  "info@example.com",
		"contact_phone":  "+1 (555) 123-4567",
		"contact_address": "123 Main St, City, State 12345",
	}

	tmpl, err := env.ParseFile("contact.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	result, err := tmpl.ExecuteToString(context)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	fmt.Println(result)
}

func testTemplateCaching(env *runtime.Environment) {
	fmt.Printf("Cache size before: %d\n", env.CacheSize())

	// Parse the same template multiple times
	for i := 0; i < 3; i++ {
		_, err := env.ParseFile("page.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			return
		}
	}

	fmt.Printf("Cache size after: %d\n", env.CacheSize())
	fmt.Println("Templates are cached successfully!")

	// Clear cache and verify
	env.ClearCache()
	fmt.Printf("Cache size after clear: %d\n", env.CacheSize())
}