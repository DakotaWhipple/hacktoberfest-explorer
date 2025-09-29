# 🎃 Hacktoberfest Repository & Issue Explorer

A powerful CLI tool to discover relevant Hacktoberfest repositories and issues tailored to your skills and interests.

## Features

- 🔍 **Smart Repository Search**: Find Hacktoberfest repos with minimum star requirements
- 🎯 **Relevance Scoring**: Repositories ranked by stars, language match, and activity
- 🐛 **Issue Discovery**: Browse beginner-friendly issues with difficulty assessment
- ⌨️ **Arrow Key Navigation**: Intuitive keyboard controls for browsing
- 🎨 **Rich Information Display**: Color-coded, well-formatted information at every step
- ⚙️ **Configurable Preferences**: Customize languages, skill level, and search limits

## Installation

1. **Clone or set up the project:**
   ```bash
   # If you haven't already
   git clone <your-repo>
   cd hacktober
   ```

2. **Build the application:**
   ```bash
   go build -o hacktober ./cmd/hacktober
   ```

## Configuration

### Method 1: Environment Variable
```bash
export GITHUB_TOKEN="your_token_here"
./hacktober
```

### Method 2: Configuration File
1. Copy the example config:
   ```bash
   cp .hacktober-config.example.json ~/.hacktober-config.json
   ```

2. Edit `~/.hacktober-config.json` with your preferences:
   ```json
   {
     "github_token": "your_github_token_here",
     "preferred_languages": ["Go", "JavaScript", "Python"],
     "skill_level": "intermediate",
     "max_repos": 30,
     "max_issues_per_repo": 15
   }
   ```

### Getting a GitHub Token

1. Go to [GitHub Settings > Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select these scopes:
   - `public_repo` - Access public repositories
   - `read:org` - Read organization data
4. Copy the generated token to your config

## Usage

### Starting the Application
```bash
./hacktober
```

### Navigation Controls

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate up/down in lists |
| `←`/`→` | Previous/next page |
| `Enter` | Select item or advance to next screen |
| `Q`/`Esc` | Go back or quit |
| `R` | Refresh current data |

### Screen Flow

1. **Welcome Screen**: Overview of your configuration
2. **Repository Search**: Displays search progress  
3. **Repository List**: Browse Hacktoberfest repos with:
   - Repository name and owner
   - Star count and primary language
   - Relevance score
   - Description and last update
4. **Issue List**: View issues in selected repo with:
   - Issue title and number
   - Difficulty assessment (Easy/Medium/Hard/Expert)
   - Labels and comment count
   - Creation date
5. **Issue Details**: Full issue information including:
   - Complete description
   - Author and metadata
   - Direct GitHub URL

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `github_token` | Your GitHub personal access token | **Required** |
| `preferred_languages` | Languages you want to work with | `["Go", "JavaScript", "Python", "TypeScript"]` |
| `skill_level` | Your experience level | `"intermediate"` |
| `max_repos` | Maximum repositories to fetch | `50` |
| `max_issues_per_repo` | Maximum issues per repository | `20` |

## How It Works

### Repository Scoring
Repositories are scored based on:
- **Star count**: More stars = higher relevance
- **Language match**: Matches your preferred languages  
- **Recent activity**: Recently updated repos score higher
- **Hacktoberfest participation**: Must have `hacktoberfest` topic

### Issue Difficulty Assessment
Issues are automatically categorized by:
- **Labels**: "good first issue", "help wanted", "beginner", etc.
- **Comment activity**: More discussion might indicate complexity
- **Issue type**: Bugs vs features vs documentation

### Search Filters
- Minimum 20 stars (configurable)
- Must have `hacktoberfest` topic
- Filters by your preferred languages
- Prioritizes "good first issue" and "help wanted" labels

## Tips for Success

1. **Start with Easy Issues**: Look for green "Easy" difficulty tags
2. **Read Descriptions Carefully**: Use the detailed view to understand requirements
3. **Check Recent Activity**: Active repositories are more likely to review PRs quickly
4. **Match Your Skills**: The tool prioritizes your preferred languages
5. **Use Multiple Languages**: Expand your language list for more opportunities

## Troubleshooting

### "GitHub token not found"
- Ensure your token is set via environment variable or config file
- Check that the token has the correct scopes

### "No repositories found"
- Try expanding your `preferred_languages` list
- Lower the star requirement threshold
- Check if Hacktoberfest is currently active

### Navigation issues
- Ensure your terminal supports ANSI colors and cursor movement
- Try running in a different terminal if arrow keys don't work

## Contributing

This tool itself could be a great Hacktoberfest project! Feel free to:
- Report bugs
- Suggest new features  
- Improve the user interface
- Add new filtering options
- Enhance the scoring algorithms

---

Happy Hacking! 🎃👩‍💻👨‍💻