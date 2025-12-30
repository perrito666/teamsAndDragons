package storage

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"teamsAndDragons/internal/domain"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// HTMLExporter exports tracker data to HTML.
type HTMLExporter struct {
	storage Storage
	md      goldmark.Markdown
}

// NewHTMLExporter creates a new HTML exporter.
func NewHTMLExporter(storage Storage) *HTMLExporter {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	return &HTMLExporter{
		storage: storage,
		md:      md,
	}
}

// PersonExportData contains all data for exporting a person.
type PersonExportData struct {
	Person     domain.Person
	Milestones []MilestoneExportData
	ExportedAt time.Time
}

// MilestoneExportData contains milestone data with computed fields.
type MilestoneExportData struct {
	domain.Milestone
	PositiveCount int
	NegativeCount int
	AchievedCount int
	PendingCount  int
}

// ExportPerson exports a single person's data to HTML.
func (e *HTMLExporter) ExportPerson(personID string, opts ExportOptions) error {
	person, err := e.storage.GetPerson(personID)
	if err != nil {
		return err
	}

	milestones, err := e.storage.ListMilestones(personID)
	if err != nil {
		return err
	}

	data := PersonExportData{
		Person:     *person,
		ExportedAt: time.Now(),
	}

	for _, m := range milestones {
		mData := MilestoneExportData{
			Milestone:     m,
			PositiveCount: len(m.PositiveActions()),
			NegativeCount: len(m.NegativeActions()),
			AchievedCount: len(m.AchievedObjectives()),
			PendingCount:  len(m.PendingObjectives()),
		}
		data.Milestones = append(data.Milestones, mData)
	}

	html, err := e.renderPersonHTML(data, opts.IncludeStyles)
	if err != nil {
		return err
	}

	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s-export.html", person.Slug())
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("writing HTML file: %w", err)
	}

	return nil
}

// ExportAll exports all people to HTML files in a directory.
func (e *HTMLExporter) ExportAll(opts ExportOptions) error {
	people, err := e.storage.ListPeople()
	if err != nil {
		return err
	}

	for _, person := range people {
		personOpts := opts
		if opts.OutputPath != "" {
			personOpts.OutputPath = filepath.Join(opts.OutputPath, person.Slug()+".html")
		}
		if err := e.ExportPerson(person.ID, personOpts); err != nil {
			return fmt.Errorf("exporting %s: %w", person.Name, err)
		}
	}

	return nil
}

func (e *HTMLExporter) renderPersonHTML(data PersonExportData, includeStyles bool) ([]byte, error) {
	tmpl, err := template.New("person").Funcs(template.FuncMap{
		"renderMarkdown": func(text string) template.HTML {
			var buf bytes.Buffer
			if err := e.md.Convert([]byte(text), &buf); err != nil {
				return template.HTML(template.HTMLEscapeString(text))
			}
			return template.HTML(buf.String())
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("January 2, 2006")
		},
		"impactClass": func(impact domain.Impact) string {
			if impact == domain.ImpactPositive {
				return "positive"
			}
			return "negative"
		},
		"impactIcon": func(impact domain.Impact) string {
			if impact == domain.ImpactPositive {
				return "+"
			}
			return "-"
		},
		"achievedClass": func(achieved bool) string {
			if achieved {
				return "achieved"
			}
			return "pending"
		},
	}).Parse(personHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	templateData := struct {
		PersonExportData
		IncludeStyles bool
	}{
		PersonExportData: data,
		IncludeStyles:    includeStyles,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}

const personHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Person.Name}} - Career Progression</title>
    {{if .IncludeStyles}}
    <style>
        :root {
            --positive: #22c55e;
            --negative: #ef4444;
            --achieved: #3b82f6;
            --pending: #9ca3af;
            --bg: #ffffff;
            --text: #1f2937;
            --border: #e5e7eb;
            --card-bg: #f9fafb;
        }
        * { box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 2rem;
            background: var(--bg);
            color: var(--text);
        }
        h1 { border-bottom: 2px solid var(--border); padding-bottom: 0.5rem; }
        h2 { color: #374151; margin-top: 2rem; }
        h3 { color: #4b5563; }
        .meta { color: #6b7280; font-size: 0.875rem; margin-bottom: 1rem; }
        .milestone {
            background: var(--card-bg);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 1.5rem;
            margin: 1.5rem 0;
        }
        .milestone-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 1rem;
        }
        .status {
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
        }
        .status.in_progress { background: #fef3c7; color: #92400e; }
        .status.completed { background: #d1fae5; color: #065f46; }
        .stats {
            display: flex;
            gap: 1rem;
            flex-wrap: wrap;
            margin: 1rem 0;
        }
        .stat {
            padding: 0.5rem 1rem;
            background: white;
            border: 1px solid var(--border);
            border-radius: 6px;
            font-size: 0.875rem;
        }
        .stat.positive { border-left: 3px solid var(--positive); }
        .stat.negative { border-left: 3px solid var(--negative); }
        .stat.achieved { border-left: 3px solid var(--achieved); }
        .objective {
            margin: 1rem 0;
            padding: 1rem;
            background: white;
            border-radius: 6px;
            border: 1px solid var(--border);
        }
        .objective-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .badge {
            padding: 0.125rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 500;
        }
        .badge.achieved { background: #dbeafe; color: #1e40af; }
        .badge.pending { background: #f3f4f6; color: #6b7280; }
        .review-notes {
            background: #fffbeb;
            border-left: 3px solid #f59e0b;
            padding: 0.75rem;
            margin: 0.5rem 0;
            font-style: italic;
        }
        .actions { margin-top: 1rem; }
        .action {
            display: flex;
            gap: 0.75rem;
            padding: 0.5rem 0;
            border-bottom: 1px solid var(--border);
        }
        .action:last-child { border-bottom: none; }
        .action-icon {
            width: 24px;
            height: 24px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
            flex-shrink: 0;
        }
        .action-icon.positive { background: #dcfce7; color: var(--positive); }
        .action-icon.negative { background: #fee2e2; color: var(--negative); }
        .action-date { color: #6b7280; font-size: 0.875rem; }
        .action-notes { color: #6b7280; font-size: 0.875rem; margin-top: 0.25rem; }
        .progress-bar {
            height: 8px;
            background: var(--border);
            border-radius: 4px;
            overflow: hidden;
            margin: 0.5rem 0;
        }
        .progress-fill {
            height: 100%;
            background: var(--achieved);
            transition: width 0.3s;
        }
        footer {
            margin-top: 3rem;
            padding-top: 1rem;
            border-top: 1px solid var(--border);
            color: #9ca3af;
            font-size: 0.75rem;
        }
    </style>
    {{end}}
</head>
<body>
    <h1>{{.Person.Name}}</h1>
    <div class="meta">
        Last updated: {{formatDate .Person.UpdatedAt}}
    </div>

    {{if .Person.Notes}}
    <div class="notes">
        {{renderMarkdown .Person.Notes}}
    </div>
    {{end}}

    <h2>Milestones</h2>
    {{range .Milestones}}
    <div class="milestone">
        <div class="milestone-header">
            <h3>{{.Number}}. {{.Name}}</h3>
            <span class="status {{.Status}}">{{.Status}}</span>
        </div>

        {{if .Description}}
        <div class="description">
            {{renderMarkdown .Description}}
        </div>
        {{end}}

        <div class="stats">
            <div class="stat positive">+{{.PositiveCount}} positive actions</div>
            <div class="stat negative">-{{.NegativeCount}} negative actions</div>
            <div class="stat achieved">{{.AchievedCount}}/{{len .Objectives}} objectives achieved</div>
        </div>

        <div class="progress-bar">
            <div class="progress-fill" style="width: {{.Progress}}%"></div>
        </div>

        {{range .Objectives}}
        <div class="objective">
            <div class="objective-header">
                <strong>{{.Name}}</strong>
                <span class="badge {{achievedClass .Achieved}}">
                    {{if .Achieved}}Achieved{{else}}Pending{{end}}
                </span>
            </div>

            {{if .Description}}
            <p>{{.Description}}</p>
            {{end}}

            {{if .ReviewNotes}}
            <div class="review-notes">{{.ReviewNotes}}</div>
            {{end}}

            {{if .Actions}}
            <div class="actions">
                {{range .Actions}}
                <div class="action">
                    <div class="action-icon {{impactClass .Impact}}">{{impactIcon .Impact}}</div>
                    <div>
                        <div><strong>{{.Description}}</strong></div>
                        <div class="action-date">{{formatDate .Date}}</div>
                        {{if .Notes}}<div class="action-notes">{{.Notes}}</div>{{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
    </div>
    {{end}}

    <footer>
        Exported on {{formatDate .ExportedAt}} | Career Progression Tracker
    </footer>
</body>
</html>
`

// Ensure HTMLExporter implements Exporter interface.
var _ Exporter = (*HTMLExporter)(nil)
