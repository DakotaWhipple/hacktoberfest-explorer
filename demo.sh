#!/bin/bash

# Hacktoberfest CLI Demo Script
# This script demonstrates how to use the hacktoberfest CLI tool

echo "🎃 Hacktoberfest Repository & Issue Explorer Demo"
echo "=================================================="
echo

echo "📋 This CLI tool helps you:"
echo "  • Find relevant Hacktoberfest repositories (20+ stars)"
echo "  • Browse beginner-friendly issues"
echo "  • Navigate with arrow keys"
echo "  • View detailed issue information"
echo

echo "⚙️  To use this tool, you need:"
echo "  1. A GitHub personal access token"
echo "  2. Set it via: export GITHUB_TOKEN='your_token_here'"
echo "  3. Or create ~/.hacktober-config.json"
echo

echo "🚀 Example configuration (~/.hacktober-config.json):"
cat .hacktober-config.example.json
echo

echo "🔧 Build the tool:"
echo "   go build -o hacktober ./cmd/hacktober"
echo

echo "🎯 Run the tool:"
echo "   GITHUB_TOKEN=your_token ./hacktober"
echo

echo "📖 Navigation controls:"
echo "   ↑/↓  - Navigate lists"
echo "   ←/→  - Change pages"
echo "   Enter - Select/Next screen"
echo "   Q/Esc - Back/Quit"
echo "   R     - Refresh data"
echo

echo "🎉 Happy Hacking!"