package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		record, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'signup'",
			map[string]any{},
		)
		if record == nil {
			record = core.NewRecord(settings)
			record.Set("name", "signup")
		}

		record.Set("data", map[string]any{
			"title":        "Candidatura Realmen",
			"submit_label": "Invia richiesta",
			"fields": []map[string]any{
				{
					"key":      "name",
					"type":     "text",
					"required": true,
					"label":    "Nome e cognome",
				},
				{
					"key":      "email",
					"type":     "email",
					"required": true,
					"label":    "Email",
				},
				{
					"key":      "mobile",
					"type":     "text",
					"required": true,
					"label":    "Cellulare",
				},
				{
					"key":      "region",
					"type":     "select",
					"required": true,
					"label":    "Regione",
					"options": []string{
						"Abruzzo",
						"Basilicata",
						"Calabria",
						"Campania",
						"Emilia-Romagna",
						"Friuli-Venezia Giulia",
						"Lazio",
						"Liguria",
						"Lombardia",
						"Marche",
						"Molise",
						"Piemonte",
						"Puglia",
						"Sardegna",
						"Sicilia",
						"Toscana",
						"Trentino-Alto Adige",
						"Umbria",
						"Valle d'Aosta",
						"Veneto",
						"Estero:input",
					},
				},
				{
					"key":      "birth_year",
					"type":     "select",
					"required": true,
					"label":    "Anno di nascita",
					"options": []string{
						"1980", "1981", "1982", "1983", "1984", "1985", "1986", "1987", "1988", "1989",
						"1990", "1991", "1992", "1993", "1994", "1995", "1996", "1997", "1998", "1999",
						"2000", "2001", "2002", "2003", "2004", "2005",
						"Altro:input",
					},
				},
				{
					"key":      "marital_status",
					"type":     "select",
					"required": true,
					"label":    "Stato relazionale",
					"options": []string{
						"Single",
						"Fidanzato",
						"Sposato",
						"Convivente",
						"Separato",
						"Vedovo",
					},
				},
				{
					"key":      "children",
					"type":     "select",
					"required": true,
					"label":    "Figli",
					"options": []string{
						"No",
						"1",
						"2",
						"3",
						"4",
						"5+",
					},
				},
				{
					"key":      "motivation",
					"type":     "textarea",
					"required": true,
					"label":    "Motivazione",
				},
			},
			"request_defaults": map[string]any{
				"status":   "1-submitted",
				"rejected": false,
			},
		})

		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}
