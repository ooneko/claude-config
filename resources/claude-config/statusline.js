#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

function getContextUsage(context) {
    try {
        // Check if context contains token information
        if (context.context_tokens_used && context.context_tokens_limit) {
            const used = context.context_tokens_used;
            const limit = context.context_tokens_limit;
            if (limit > 0) {
                const percentage = (used / limit) * 100;
                return `${Math.round(percentage)}%`;
            }
        }
        
        // Check if context has tokens_until_compact (inverse calculation)
        if (context.tokens_until_compact && context.context_tokens_limit) {
            const remaining = context.tokens_until_compact;
            const limit = context.context_tokens_limit;
            if (limit > 0) {
                const percentage = ((limit - remaining) / limit) * 100;
                return `${Math.round(percentage)}%`;
            }
        }
        
        // Check if context directly provides percentage
        if (context.context_percentage) {
            return `${Math.round(context.context_percentage)}%`;
        }
        
        // Check for alternative field names
        if (context.context && typeof context.context === 'object') {
            const ctx = context.context;
            if (ctx.usage) {
                return `${Math.round(ctx.usage)}%`;
            }
            if (ctx.percentage) {
                return `${Math.round(ctx.percentage)}%`;
            }
        }
        
        // Try to estimate based on Claude's internal calculation
        return estimateContextFromSession(context);
    } catch (e) {
        return "N/A";
    }
}

function estimateContextFromSession(context) {
    try {
        // Constants from obfuscated code
        const YE8 = 13000;  // Auto-compact threshold offset
        const modelLimits = {
            'opus-4': 200000,
            'opus-4-1': 200000,
            'sonnet-4': 200000,
            'haiku': 200000,
            'default': 200000
        };
        
        // Get model name
        const modelName = (context.model?.display_name || 'default').toLowerCase();
        
        // Find matching limit
        let maxTokens = modelLimits.default;
        for (const [key, limit] of Object.entries(modelLimits)) {
            if (modelName.includes(key)) {
                maxTokens = limit;
                break;
            }
        }
        
        let tokenUsage = null;
        
        // Try to read from transcript file if available
        if (context.transcript_path) {
            try {
                const transcriptPath = context.transcript_path;
                if (fs.existsSync(transcriptPath)) {
                    const content = fs.readFileSync(transcriptPath, 'utf8');
                    const lines = content.split('\n').filter(line => line.trim());
                    
                    let messageCount = 0;
                    let totalChars = 0;
                    
                    for (const line of lines) {
                        if (line.trim()) {
                            messageCount++;
                            try {
                                const msg = JSON.parse(line);
                                if (typeof msg === 'object') {
                                    const content = String(msg.content || '');
                                    totalChars += content.length;
                                }
                            } catch {
                                // Fallback: estimate 100 chars per message
                                totalChars += 100;
                            }
                        }
                    }
                    
                    // Estimate tokens (roughly 4 chars per token)
                    tokenUsage = Math.floor(totalChars / 4);
                    
                    // Alternative: use message count if char count seems wrong
                    if (tokenUsage < messageCount * 10) {
                        tokenUsage = messageCount * 100;
                    }
                }
            } catch (e) {
                tokenUsage = null;
            }
        }
        
        // Fall back to other estimation methods if transcript reading failed
        if (tokenUsage === null) {
            if (context.session_tokens) {
                tokenUsage = context.session_tokens;
            } else if (context.message_count) {
                tokenUsage = context.message_count * 100;
            } else if (context.conversation_length) {
                tokenUsage = Math.floor(context.conversation_length / 4);
            } else {
                return "N/A";
            }
        }
        
        // Calculate percentage using Claude's formula
        const threshold = maxTokens - YE8;  // 200000 - 13000 = 187000
        const percentUsed = Math.round((tokenUsage / threshold) * 100);
        
        // Ensure it doesn't exceed 100%
        const finalPercent = Math.min(100, Math.max(0, percentUsed));
        
        return `${finalPercent}%`;
    } catch {
        return "N/A";
    }
}

function getCcusageCosts() {
    try {
        const today = new Date();
        // Use local date instead of UTC to avoid timezone issues
        const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
        
        // Calculate date ranges using local time
        const weekStart = new Date(today.getTime() - 6 * 24 * 60 * 60 * 1000);
        const weekStartStr = `${weekStart.getFullYear()}-${String(weekStart.getMonth() + 1).padStart(2, '0')}-${String(weekStart.getDate()).padStart(2, '0')}`;
        
        const monthStart = new Date(today.getTime() - 29 * 24 * 60 * 60 * 1000);
        const monthStartStr = `${monthStart.getFullYear()}-${String(monthStart.getMonth() + 1).padStart(2, '0')}-${String(monthStart.getDate()).padStart(2, '0')}`;
        
        // Run ccusage daily command
        const result = execSync('ccusage daily --limit 30', { 
            encoding: 'utf8', 
            timeout: 5000 
        });
        
        const lines = result.split('\n');
        let currentYear = null;
        let dailyCost = 0.00;
        let weeklyCost = 0.00;
        let monthlyCost = 0.00;
        
        for (let i = 0; i < lines.length; i++) {
            // Strip ANSI color codes
            const cleanLine = lines[i].replace(/\x1b\[[0-9;]*m/g, '');
            
            // Check if this line has a year
            const yearMatch = cleanLine.match(/│\s*(\d{4})\s/);
            if (yearMatch) {
                currentYear = yearMatch[1];
                // Check if there's a cost on this line - revert to working regex
                const costMatch = cleanLine.match(/\$\s*(\d+(?:,\d{3})*(?:\.\d{2})?)\s*│/);
                if (costMatch) {
                    const costStr = costMatch[1].replace(/,/g, '');
                    const cost = parseFloat(costStr);
                    
                    // Check the next line for month-day
                    if (i + 1 < lines.length) {
                        const nextLine = lines[i + 1];
                        const cleanNext = nextLine.replace(/\x1b\[[0-9;]*m/g, '');
                        const dateMatch = cleanNext.match(/│\s*(\d{2})-(\d{2})\s/);
                        if (dateMatch && currentYear) {
                            const month = dateMatch[1];
                            const day = dateMatch[2];
                            const dateStr = `${currentYear}-${month}-${day}`;
                            
                            // Check if this date is today
                            if (dateStr === todayStr) {
                                dailyCost = cost;
                            }
                            
                            // Check if this date is within ranges
                            if (dateStr >= weekStartStr && dateStr <= todayStr) {
                                weeklyCost += cost;
                            }
                            
                            if (dateStr >= monthStartStr && dateStr <= todayStr) {
                                monthlyCost += cost;
                            }
                        }
                    }
                }
            }
        }
        
        return `D:$${dailyCost.toFixed(2)} W:$${weeklyCost.toFixed(2)} M:$${monthlyCost.toFixed(2)}`;
    } catch (e) {
        console.error(`Error getting ccusage costs: ${e}`, { file: process.stderr });
        return "D:$0.00 W:$0.00 M:$0.00";
    }
}

function generateStatusLine(context) {
    // Extract information from context
    const model = context.model?.display_name || 'Unknown';
    const currentDir = context.workspace?.current_directory || process.cwd();
    const projectName = path.basename(currentDir);
    
    // Get context usage
    const contextUsage = getContextUsage(context);
    
    // Get ccusage costs
    const costs = getCcusageCosts();
    
    // ANSI color codes
    const BLUE = '\033[94m';      // Blue for project
    const GREEN = '\033[92m';     // Green for model
    const YELLOW = '\033[93m';    // Yellow for context
    const CYAN = '\033[96m';      // Cyan for costs
    const RESET = '\033[0m';      // Reset color
    const BOLD = '\033[1m';       // Bold text
    
    // Format the status line with colors
    const statusLine = 
        `${BLUE}${BOLD}📁 ${projectName}${RESET} | ` +
        `${GREEN}${BOLD}🤖 ${model}${RESET} | ` +
        `${YELLOW}${BOLD}📊 ${contextUsage}${RESET} | ` +
        `${CYAN}${BOLD}💰 ${costs}${RESET}`;
    
    return statusLine;
}

// Main execution
try {
    let context;
    
    if (process.argv.length > 2) {
        // Read JSON input from command line argument
        context = JSON.parse(process.argv[2]);
    } else {
        // Read from stdin if no argument provided
        const input = require('fs').readFileSync(0, 'utf-8');
        context = JSON.parse(input);
    }
    
    // Generate and print the status line
    console.log(generateStatusLine(context));
} catch (e) {
    // Fallback status line if something goes wrong
    console.log(`Status Line Error: ${e.toString().substring(0, 50)}`);
}