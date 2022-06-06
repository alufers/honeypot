package main

import (
	"embed"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/websocket/v2"
)

//go:embed honeypot-frontend/public
var frontendStatic embed.FS

func RunAdminServer() {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(frontendStatic),
		PathPrefix: "honeypot-frontend/public",
		Browse:     true,
	}))
	// server files from frontend folder
	app.Static("/", "./frontend")
	app.Get("/api/protocols", func(c *fiber.Ctx) error {
		listeningProtocolsMutex.Lock()
		defer listeningProtocolsMutex.Unlock()
		return c.JSON(listeningProtocols)
	})

	app.Get("/api/attacks", func(c *fiber.Ctx) error {
		query := struct {
			Count           string   `query:"count"`
			Before          string   `query:"before"`
			Classifications []string `query:"classifications"`
			Protocols       []string `query:"protocols"`
		}{}
		if err := c.QueryParser(&query); err != nil {
			return err
		}
		count, _ := strconv.Atoi(query.Count)
		if count == 0 {
			count = 10
		}
		before, _ := strconv.Atoi(query.Before)
		if before == 0 {
			before = math.MaxInt
		}

		var attacks = make([]*Attack, 0)
		dbQuery := db.Where("id < ?", before)
		if query.Classifications != nil && len(query.Classifications) > 0 {
			dbQuery = dbQuery.Where("classification IN (?)", query.Classifications)
		}
		if query.Protocols != nil && len(query.Protocols) > 0 {
			dbQuery = dbQuery.Where("protocol IN (?)", query.Protocols)
		}
		if err := dbQuery.Order("id desc").Limit(count).Find(&attacks).Error; err != nil {
			log.Printf("failed to get attacks: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		log.Printf("debug: %v", attacks)
		currentAttacksMutex.Lock()
		defer currentAttacksMutex.Unlock()
		for i, a := range attacks {
			if attack, ok := currentAttacks[a.ID]; ok {
				attacks[i] = attack
			}
		}
		return c.JSON(attacks)

	})

	app.Get("/api/attacks/stats/by-country", func(c *fiber.Ctx) error {

		stats := make([]map[string]interface{}, 0)
		if err := db.Raw("SELECT country, country_code, count(*) AS count FROM attacks GROUP BY country,country_code ORDER BY count DESC").Scan(&stats).Error; err != nil {
			log.Printf("failed to get attack stats: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	app.Get("/api/attacks/stats/by-time", func(c *fiber.Ctx) error {

		stats := make([]map[string]interface{}, 0)
		if err := db.Raw("SELECT  strftime(?, created_at) as time, protocol, count(*) AS count FROM attacks GROUP BY protocol, time").Scan(&stats).Error; err != nil {
			log.Printf("failed to get attack stats: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	app.Get("/api/credentials/stats/passwords", func(c *fiber.Ctx) error {

		stats := make([]map[string]interface{}, 0)
		if err := db.Raw("SELECT password, count(*) AS count, 100.0 * COUNT(*) / (SELECT COUNT(*) FROM credential_usages) AS percentage FROM credential_usages GROUP BY password ORDER BY count DESC LIMIT 20;").Scan(&stats).Error; err != nil {
			log.Printf("failed to get attack stats: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	app.Get("/api/attacks/ws", websocket.New(func(c *websocket.Conn) {
		var (
			err error
		)
		listener := make(chan *Attack)
		cancel := attacksEventBroadcaster.AddListener(listener)
		defer cancel()
		for {
			at := <-listener
			var jsonData []byte
			jsonData, _ = json.Marshal(at)
			if err = c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Println("websocket write:", err)
				break
			}
		}
	}))

	log.Fatal(app.Listen(getEnv("ADMIN_ADDR", "localhost:7878")))
}
