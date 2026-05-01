package config

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Guild  string `yaml:"guild"`
	Index  Index  `yaml:"index"`
	Label  string `yaml:"label"`
	Invite string `yaml:"invite"`
}

type Index int

type Servers struct {
	servers map[Index]Server
}

func LoadServers(path string) *Servers {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("servers: failed to read config: %v", err)
		return &Servers{servers: map[Index]Server{}}
	}

	var raw struct {
		Servers []Server `yaml:"servers"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Printf("servers: failed to parse config: %v", err)
		return &Servers{servers: map[Index]Server{}}
	}

	m := make(map[Index]Server, len(raw.Servers))
	for _, s := range raw.Servers {
		m[s.Index] = s
	}

	log.Printf("loaded servers config with %d entries", len(m))
	return &Servers{servers: m}
}

func (s *Servers) FindByID(id Index) *Server {
	if server, ok := s.servers[id]; ok {
		return &server
	}
	return nil
}

func (s *Servers) FindByName(q string) []*Server {
	q = strings.ToLower(q)
	var results []*Server
	for _, server := range s.servers {
		if strings.Contains(strings.ToLower(server.Label), q) {
			srv := server
			results = append(results, &srv)
		}
	}
	return results
}

func (s *Servers) All() []*Server {
	results := make([]*Server, 0, len(s.servers))
	for _, server := range s.servers {
		srv := server
		results = append(results, &srv)
	}
	return results
}
