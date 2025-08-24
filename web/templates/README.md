# Web Templates

This directory contains Templ templates for the Go Templ Template project.

## Structure

```
web/templates/
├── layouts/          # Layout templates
│   └── base.templ   # Base HTML layout with header/footer
├── components/       # Reusable components
│   ├── header.templ # Site header with navigation
│   ├── footer.templ # Site footer with links
│   └── navigation.templ # Navigation components
└── pages/           # Page templates
    └── home.templ   # Homepage template
```

## Components

### Base Layout (`layouts/base.templ`)
- Responsive HTML5 layout
- Includes meta tags for SEO and mobile
- Integrates Tailwind CSS
- Includes header, main content area, and footer
- Mobile-first responsive design

### Header (`components/header.templ`)
- Responsive navigation bar
- Logo and brand name
- Desktop and mobile navigation menus
- Authentication buttons (Login/Sign Up)
- Mobile hamburger menu with JavaScript toggle

### Footer (`components/footer.templ`)
- Multi-column responsive footer
- Brand information and description
- Quick links and support links
- Social media icons
- Copyright notice with dynamic year

### Navigation (`components/navigation.templ`)
- Reusable navigation component
- Support for active states
- Desktop and mobile variants
- Configurable navigation items

## Features

### Responsive Design
- Mobile-first approach using Tailwind CSS
- Breakpoints: sm (640px), md (768px), lg (1024px), xl (1280px)
- Flexible grid layouts and responsive typography

### Accessibility
- Semantic HTML structure
- ARIA labels for interactive elements
- Keyboard navigation support
- Screen reader friendly

### Performance
- Minimal JavaScript for mobile menu
- Optimized CSS with Tailwind's utility classes
- Fast rendering with Templ's compiled templates

## Usage

### Basic Page Template
```go
package pages

import "go-templ-template/web/templates/layouts"

templ MyPage() {
    @layouts.Base("Page Title", myPageContent())
}

templ myPageContent() {
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <h1 class="text-3xl font-bold text-gray-900">My Page</h1>
        <p class="text-gray-600 mt-4">Page content goes here.</p>
    </div>
}
```

### Using Navigation Component
```go
import "go-templ-template/web/templates/components"

navItems := []components.NavItem{
    {Label: "Home", URL: "/", Active: true},
    {Label: "About", URL: "/about", Active: false},
}

@components.Navigation(navItems)
```

## Development

### Generate Templates
```bash
templ generate
```

### Build CSS
```bash
npm run build-css
```

### Run Tests
```bash
go test ./web/templates/...
```

### Hot Reload
```bash
air
```

## Styling

The templates use Tailwind CSS with custom component classes defined in `web/static/css/input.css`:

- `.btn-primary` - Primary button styling
- `.btn-secondary` - Secondary button styling  
- `.card` - Card component styling
- `.form-input` - Form input styling
- `.form-label` - Form label styling

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile browsers (iOS Safari, Chrome Mobile)
- Progressive enhancement for older browsers