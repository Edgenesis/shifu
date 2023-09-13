package prompts

const (
	GreetingPrompts = `I will give you two templates, please note.`

	TemplateENPrompts = `
	# Changelog since [v0.x.0](https://github.com/Edgenesis/shifu/releases/tag/v0.x.0)

	## New Features 🎉
	
	## Bug fixes
	
	## Enhancement
	
	## Docs
	
	## New Contributors 🌟
	
	## Dependabot 🤖
	
	**Full Changelog**: https://github.com/Edgenesis/shifu/compare/v0.x.0...v0.y.0`

	TemplateZHPrompts = `
	# 自 [v0.x.0](https://github.com/Edgenesis/shifu/releases/tag/v0.x.0) 以来的变更

	## 新功能 🎉
	
	## Bug 修复
	
	## 改进
	
	## 文档
	
	## 新的贡献者 🌟
	
	## Dependabot 自动更新 🤖
	
	**完整变更日志**: https://github.com/Edgenesis/shifu/compare/v0.x.0...v0.y.0`

	GeneratePrompts = `
	Then I will give you a json formatted response.
	Please generate two markdown files according to the two templates I provided.
	One is English version and the other is Chinese version, please translate the neccessary words to Chinese as well.
	And please OMIT the EMPTY fields.
	divide each version by '--------'
	Your answer MUST not contain any other contents unrelative to the md, which means you are only allowed to output markdown.
	`
)
