package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"poe2campain/internal/campaign"
	appconfig "poe2campain/internal/config"
	"poe2campain/internal/importer"
	"poe2campain/internal/logreader"
	appsession "poe2campain/internal/runtime"
)

const defaultDataPath = "data/campaign.normalized.json"
const defaultSourcesPath = "data/sources"

func main() {
	log.SetFlags(0)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "--help", "-h":
			printUsage()
			return
		case "update-data", "generate-data":
			updateData(os.Args[2:])
			return
		case "validate-data":
			validateData(os.Args[2:])
			return
		case "config":
			configCommand(os.Args[2:])
			return
		case "state":
			stateCommand(os.Args[2:])
			return
		}
	}

	run(os.Args[1:])
}

func printUsage() {
	fmt.Println(`Usage: poe2campain [command] [flags]
Commands:
  help              show this help
  update-data       regenerate normalized campaign data from sources
  validate-data     validate campaign data
  config            manage config (path, show, init, set-client)
  state             manage session state (path, show, reset)
Flags:
  -client string    Path of Exile 2 Client.txt path (auto-detected if empty)
  -config string    config file path (default ` + appconfig.DefaultPathDisplay() + `)
  -data string      normalized campaign data path (default "` + defaultDataPath + `")
  -debug-client     scan Client.txt and show the latest matched area
  -debug-zone string  match and render route entry for a zone or area ID
  -list-zones       list known zones`)
}

func updateData(args []string) {
	fs := flag.NewFlagSet("update-data", flag.ExitOnError)
	sourcesPath := fs.String("sources", defaultSourcesPath, "source snapshot directory")
	outPath := fs.String("out", defaultDataPath, "normalized data output path")
	_ = fs.Parse(args)

	if err := importer.GenerateNormalizedFromDomistae(*sourcesPath, *outPath); err != nil {
		log.Fatal(err)
	}
	_, report, err := campaign.Load(*outPath)
	if err != nil {
		log.Fatal(err)
	}
	printReport(report)
	fmt.Println("Output:", *outPath)
}

func validateData(args []string) {
	fs := flag.NewFlagSet("validate-data", flag.ExitOnError)
	dataPath := fs.String("data", resolveDefaultDataPath(), "normalized campaign data path")
	_ = fs.Parse(args)

	_, report, err := campaign.Load(*dataPath)
	if err != nil {
		log.Fatal(err)
	}
	printReport(report)
}

func configCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Println("Usage: poe2campain config <subcommand>")
		fmt.Println("")
		fmt.Println("Subcommands:")
		fmt.Println("  path                  show config file path")
		fmt.Println("  show [--config path]  show current config")
		fmt.Println("  init [--config path] [--client path] [--force]  create config")
		fmt.Println("  set-client [--config path] /path/to/Client.txt  set Client.txt path")
		return
	}

	switch args[0] {
	case "path":
		fmt.Println(appconfig.DefaultPath())
	case "show":
		fs := flag.NewFlagSet("config show", flag.ExitOnError)
		configPath := fs.String("config", appconfig.DefaultPath(), "config file path")
		_ = fs.Parse(args[1:])
		cfg, found, err := appconfig.Load(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		if !found {
			fmt.Println("No config file found at", *configPath)
			return
		}
		printJSON(cfg)
	case "init":
		fs := flag.NewFlagSet("config init", flag.ExitOnError)
		configPath := fs.String("config", appconfig.DefaultPath(), "config file path")
		clientPath := fs.String("client", "", "Path of Exile 2 Client.txt path")
		force := fs.Bool("force", false, "overwrite existing config")
		_ = fs.Parse(args[1:])
		if appconfig.Exists(*configPath) && !*force {
			log.Fatalf("config already exists at %s (use --force to overwrite)", *configPath)
		}
		if err := appconfig.Save(*configPath, appconfig.Config{ClientPath: *clientPath}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Wrote config:", *configPath)
	case "set-client":
		fs := flag.NewFlagSet("config set-client", flag.ExitOnError)
		configPath := fs.String("config", appconfig.DefaultPath(), "config file path")
		_ = fs.Parse(args[1:])
		if fs.NArg() != 1 {
			log.Fatal("usage: poe2campain config set-client [--config path] /path/to/Client.txt")
		}
		cfg, _, err := appconfig.Load(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		cfg.ClientPath = fs.Arg(0)
		if err := appconfig.Save(*configPath, cfg); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Updated client_txt in", *configPath)
	default:
		log.Fatalf("unknown config command %q", args[0])
	}
}

func stateCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Println("Usage: poe2campain state <subcommand>")
		fmt.Println("")
		fmt.Println("Subcommands:")
		fmt.Println("  path   show state file path")
		fmt.Println("  show   show current state")
		fmt.Println("  reset  delete state file")
		return
	}
	path := appconfig.DefaultStatePath()
	switch args[0] {
	case "path":
		fmt.Println(path)
	case "show":
		snapshot, found, err := loadSessionSnapshot(path)
		if err != nil {
			log.Fatal(err)
		}
		if !found {
			fmt.Println("No state file found at", path)
			return
		}
		printJSON(snapshot)
	case "reset":
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
		fmt.Println("Reset state:", path)
	default:
		log.Fatalf("unknown state command %q", args[0])
	}
}

func run(args []string) {
	fs := flag.NewFlagSet("poe2campain", flag.ExitOnError)
	configPath := fs.String("config", "", "config file path (default "+appconfig.DefaultPathDisplay()+")")
	dataPath := fs.String("data", resolveDefaultDataPath(), "normalized campaign data path")
	listZones := fs.Bool("list-zones", false, "list known zones")
	debugZone := fs.String("debug-zone", "", "match and render route entry for a zone or generated area ID")
	debugClient := fs.Bool("debug-client", false, "scan Client.txt and show the latest matched area")
	clientPath := fs.String("client", "", "Path of Exile 2 Client.txt path (auto-detected from Steam/standalone install if empty)")
	_ = fs.Parse(args)

	cfg, _, err := appconfig.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if *clientPath == "" {
		*clientPath = cfg.ClientPath
	}
	if *clientPath == "" {
		*clientPath = appconfig.FindClientTxt()
	}

	data, _, err := campaign.Load(*dataPath)
	if err != nil {
		log.Fatal(err)
	}

	if *listZones {
		for _, zone := range data.Zones {
			fmt.Printf("%s\t%s\t%s\n", zone.ID, zone.Act, zone.Name)
		}
		return
	}
	if *debugZone != "" {
		s := appsession.New(data)
		state, ok := s.JumpToZone(*debugZone)
		if !ok {
			log.Fatalf("no zone match for %q", *debugZone)
		}
		fmt.Print(renderState(state, false, 0))
		return
	}
	if *debugClient {
		if *clientPath == "" {
			log.Fatal("--debug-client requires --client /path/to/Client.txt or config client_txt")
		}
		debugClientLog(data, *clientPath)
		return
	}
	if *clientPath == "" {
		command := os.Args[0]
		if command == "" {
			command = "poe2campain"
		}
		fmt.Println("No Client.txt found.")
		fmt.Println("")
		fmt.Println("Checked default locations:")
		for _, p := range appconfig.DefaultClientPaths() {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println("")
		fmt.Printf("Set it with: %s config set-client /path/to/Client.txt\n", command)
		fmt.Printf("  or use:    %s --client /path/to/Client.txt\n", command)
		fmt.Println("Config file:", *configPath)
		return
	}

	if err := runUI(data, *clientPath); err != nil {
		log.Fatal(err)
	}
}

type areaMsg logreader.AreaEvent
type errMsg error

type guideModel struct {
	session   *appsession.Session
	events    <-chan logreader.AreaEvent
	errs      <-chan error
	cancel    context.CancelFunc
	statePath string
	status    string
	width     int
	showHelp  bool
	quitting  bool
	err       error
}

func runUI(data *campaign.CampaignData, clientPath string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	events := make(chan logreader.AreaEvent)
	errs := make(chan error, 1)
	go func() {
		errs <- logreader.Follow(ctx, clientPath, true, 500*time.Millisecond, func(event logreader.AreaEvent) {
			select {
			case events <- event:
			case <-ctx.Done():
			}
		})
	}()

	s := appsession.New(data)
	statePath := appconfig.DefaultStatePath()
	if snapshot, found, err := loadSessionSnapshot(statePath); err != nil {
		stop()
		return err
	} else if found {
		s.Restore(snapshot)
	}

	status := "watching Client.txt"
	if event, found, err := logreader.ScanLatest(clientPath); err != nil {
		stop()
		return err
	} else if found {
		s.OnAreaDetected(event.AreaID)
		status = fmt.Sprintf("latest area: level %d %s", event.Level, event.AreaID)
	}
	if s.State().Route == nil {
		s.Start()
	}

	_, err := tea.NewProgram(guideModel{
		session:   s,
		events:    events,
		errs:      errs,
		cancel:    stop,
		statePath: statePath,
		status:    status,
	}).Run()
	return err
}

func (m guideModel) Init() tea.Cmd { return waitForClient(m.events, m.errs) }

func (m guideModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	setStatus := func(status string) {
		m.status = status
	}
	save := func() { saveSessionSnapshot(m.statePath, m.session) }
	setAndSave := func(status string) {
		setStatus(status)
		save()
	}

	switch msg := msg.(type) {
	case areaMsg:
		event := logreader.AreaEvent(msg)
		before := m.session.State().RouteIndex
		state, ok := m.session.OnAreaDetected(event.AreaID)
		if !ok {
			setStatus(fmt.Sprintf("unknown area: level %d %s", event.Level, event.AreaID))
			return m, waitForClient(m.events, m.errs)
		}
		if state.RouteIndex != before {
			setStatus(fmt.Sprintf("area: level %d %s", event.Level, event.AreaID))
		} else {
			setStatus(fmt.Sprintf("area seen: level %d %s", event.Level, event.AreaID))
		}
		return m, waitForClient(m.events, m.errs)
	case errMsg:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			m.cancel()
			save()
			return m, tea.Quit
		case "h":
			m.showHelp = !m.showHelp
			return m, nil
		case "space", " ":
			before := m.session.State()
			if after, ok := m.session.ToggleCurrentStepDoneAdvance(); ok {
				status := "marked step done"
				if before.CompletedStepIndexes[before.StepIndex] {
					status = "undid step"
				} else if after.StepIndex != before.StepIndex {
					status = "marked step done and advanced"
				}
				setAndSave(status)
			}
			return m, nil
		case "up", "k":
			m.session.PrevStep()
			setAndSave("previous step")
			return m, nil
		case "down", "j":
			m.session.NextStep()
			setAndSave("next step")
			return m, nil
		case "left":
			m.session.PrevRoute()
			setAndSave("previous route")
			return m, nil
		case "right":
			m.session.NextRoute()
			setAndSave("next route")
			return m, nil
		case "home":
			m.session.Start()
			setAndSave("start")
			return m, nil
		case "end":
			m.session.End()
			setAndSave("end")
			return m, nil
		}
	}
	return m, nil
}

func (m guideModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.err != nil {
		return tea.NewView(m.err.Error())
	}

	if m.showHelp {
		return tea.NewView(renderHelp())
	}

	v := tea.NewView(renderState(m.session.State(), true, m.width-1))
	v.AltScreen = true
	return v
}

func waitForClient(events <-chan logreader.AreaEvent, errs <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-events:
			return areaMsg(event)
		case err := <-errs:
			if err != nil {
				return errMsg(err)
			}
			return nil
		}
	}
}

func debugClientLog(data *campaign.CampaignData, clientPath string) {
	event, found, err := logreader.ScanLatest(clientPath)
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		log.Fatalf("no generated area lines found in %s", clientPath)
	}
	s := appsession.New(data)
	state, ok := s.OnAreaDetected(event.AreaID)
	if !ok {
		log.Fatalf("no zone match for %q", event.AreaID)
	}
	fmt.Printf("Latest Client.txt area: level %d area %s\n", event.Level, event.AreaID)
	fmt.Print(renderState(state, false, 0))
}

var (
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true)
	zoneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true)
	stepStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	doneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9399b2"))
)

func renderHelp() string {
	return "\n  poe2campain\n\n  ↑/k    step up\n  ↓/j    step down\n  ←      zone back\n  →      zone forward\n  space  toggle done\n  h      close help\n  q      quit\n"
}

func renderState(state appsession.State, useStyle bool, width int) string {
	if state.Route == nil {
		return "Waiting for area...\n"
	}
	var b strings.Builder
	zoneName := state.Route.DisplayName
	if state.Zone != nil {
		zoneName = state.Zone.Name
	}
	header := zoneName
	if state.Route.LevelRange != "" {
		header += "  (" + state.Route.LevelRange + ")"
	}
	if useStyle {
		b.WriteString(zoneStyle.Render(header))
	} else {
		b.WriteString(header)
	}
	b.WriteString("\n\n")

	for i, step := range state.Route.Steps {
		done := state.CompletedStepIndexes[i]
		prefix := "  "
		if i == state.StepIndex {
			prefix = "> "
		}
		if done {
			prefix += "✓ "
		}

		text := step.Text
		if step.Optional {
			text += " (optional)"
		}

		for j, part := range wrapStepText(text, width-lipgloss.Width(prefix)) {
			linePrefix := prefix
			if j > 0 {
				linePrefix = strings.Repeat(" ", lipgloss.Width(prefix))
			}
			line := linePrefix + part

			switch {
			case i == state.StepIndex && useStyle:
				line = activeStyle.Render(line)
			case done && useStyle:
				line = doneStyle.Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func wrapStepText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	return strings.Split(lipgloss.Wrap(text, width, " "), "\n")
}

func loadSessionSnapshot(path string) (appsession.Snapshot, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return appsession.Snapshot{}, false, nil
		}
		return appsession.Snapshot{}, false, err
	}
	var snapshot appsession.Snapshot
	if err := json.Unmarshal(b, &snapshot); err != nil {
		return appsession.Snapshot{}, true, fmt.Errorf("decode state %s: %w", path, err)
	}
	return snapshot, true, nil
}

func saveSessionSnapshot(path string, s *appsession.Session) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Print(err)
		return
	}
	b, err := json.MarshalIndent(s.Snapshot(), "", "  ")
	if err != nil {
		log.Print(err)
		return
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0644); err != nil {
		log.Print(err)
	}
}

func resolveDefaultDataPath() string {
	if fileExists(defaultDataPath) {
		return defaultDataPath
	}
	exe, err := os.Executable()
	if err == nil && exe != "" {
		candidate := filepath.Join(filepath.Dir(exe), defaultDataPath)
		if fileExists(candidate) {
			return candidate
		}
	}
	return defaultDataPath
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

func printReport(report campaign.ValidationReport) {
	fmt.Printf("Zones: %d\n", report.Zones)
	fmt.Printf("Route entries: %d\n", report.RouteEntries)
	fmt.Printf("Steps: %d\n", report.Steps)
	fmt.Printf("Fallback steps: %d\n", report.FallbackSteps)
	fmt.Printf("Warnings: %d\n", len(report.Warnings))
	for _, warning := range report.Warnings {
		fmt.Println("warning:", warning)
	}
}
