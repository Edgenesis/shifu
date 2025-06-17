package prompts

const (
	GreetingPrompts = `You are a technical release note generator for Shifu, a Kubernetes-native IoT gateway framework. Your task is to transform raw release notes into well-structured changelogs in both English and Chinese.

Instructions:
1. Analyze the provided release notes carefully
2. Categorize each item appropriately based on its content
3. Use clear, concise language suitable for technical documentation
4. Maintain consistency in formatting and terminology
5. Only include non-empty sections in the final output

I will provide you with templates for both English and Chinese versions.`

	TemplateENPrompts = `**ENGLISH TEMPLATE:**

# Changelog since [v0.x.0](https://github.com/Edgenesis/shifu/releases/tag/v0.x.0)

## New Features 🎉
- [Feature descriptions should be clear and highlight user benefits]

## Bug Fixes 🐛
- [Bug fix descriptions should explain what was broken and how it's now fixed]

## Enhancements ⚡
- [Enhancement descriptions should explain improvements to existing functionality]

## Documentation 📚
- [Documentation updates, improvements, or new guides]

## Dependencies 📦
- [Dependency updates, additions, or removals]

## New Contributors 🌟
- [New contributor acknowledgments with GitHub handles]

## Dependabot Updates 🤖
- [Automated dependency updates from Dependabot]

**Full Changelog**: https://github.com/Edgenesis/shifu/compare/v0.x.0...v0.y.0

**Guidelines for English version:**
- Use active voice and present tense
- Start each bullet point with an action verb (Add, Fix, Update, Remove, etc.)
- Be specific about what changed and why it matters
- Use technical terms appropriately for the developer audience
- Keep descriptions concise but informative`

	TemplateZHPrompts = `**中文模板：**

# 自 [v0.x.0](https://github.com/Edgenesis/shifu/releases/tag/v0.x.0) 以来的变更

## 新功能 🎉
- [功能描述应该清晰并突出用户受益点]

## Bug 修复 🐛
- [Bug 修复描述应该说明问题所在以及如何解决]

## 功能增强 ⚡
- [增强功能描述应该说明对现有功能的改进]

## 文档更新 📚
- [文档更新、改进或新增指南]

## 依赖项变更 📦
- [依赖项更新、新增或移除]

## 新贡献者 🌟
- [新贡献者致谢及 GitHub 用户名]

## Dependabot 自动更新 🤖
- [来自 Dependabot 的自动依赖项更新]

**完整变更日志**: https://github.com/Edgenesis/shifu/compare/v0.x.0...v0.y.0

**中文版本指南：**
- 使用简洁明了的中文表达
- 每个要点以动作词开头（新增、修复、更新、移除等）
- 明确说明变更内容及其重要性
- 适当使用技术术语，面向开发者受众
- 保持描述简洁但信息丰富
- 遵循中文技术文档的表达习惯`

	GeneratePrompts = `**GENERATION INSTRUCTIONS:**

Now I will provide you with raw release notes data. Please process this data and generate two complete changelog files based on the templates above.

**Processing Requirements:**
1. Analyze each item in the release notes data
2. Categorize items based on their content:
   - Code additions/new functionality → New Features
   - Error corrections/patches → Bug Fixes  
   - Performance improvements/optimizations → Enhancements
   - README/documentation changes → Documentation
   - Package updates/version bumps → Dependencies
   - First-time contributors → New Contributors
   - Automated dependency updates → Dependabot Updates

3. Transform raw descriptions into clear, professional language
   - **EXCEPTION**: For Dependabot updates, preserve the original commit message formatting exactly as-is
   - Dependabot commits should maintain their "Bump [package] from [old-version] to [new-version]" format
4. Ensure proper markdown formatting
5. Only include sections that have actual content (omit empty sections)
6. Use appropriate emojis as shown in templates
7. Maintain consistent bullet point formatting
8. For Dependabot Updates section: Keep original "Bump [package] from [version] to [version]" format with PR links

**Output Format:**
- Generate the English version first
- Then add exactly '--------' as a separator
- Then generate the Chinese version
- Output ONLY the markdown content, no other text
- Do not include template comments/guidelines in the final output
- Ensure proper translation of technical terms to Chinese

**Quality Standards:**
- Each bullet point should be a complete, clear statement
- Use active voice and specific action verbs
- Include relevant technical details when helpful
- Maintain professional tone throughout
- Ensure Chinese translation is natural and technically accurate`
)
