package agent

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const (
	SegmentOverviewSystemPrompt = `
	你是一个主题总结专家，你需要根据输入的Quesion-Answer对，生成一个对话主题总结，总结中需要突出关键词。

	以下是一些示例：
	单个Q&A对的情况：
	---
	Question: 你好
	Answer: 你好呀，今天过得怎么样？
	
	你的输出:
	**简单的问候**，主要内容是**打招呼**和**询问近况**
	---
	多个Q&A对的情况：
	---
	Question: 什么是Go语言？
	Answer: Go语言（或Golang）是Google开发的一种静态强类型、编译型、并发型，并具有垃圾回收功能的编程语言。
	
	Question: Go语言的主要特点是什么？
	Answer: 主要特点包括：静态类型、编译型语言、内置垃圾回收、原生支持并发（Goroutines和Channels）、以及快速的编译速度。
	
	Question: Go语言通常用于哪些场景？
	Answer: 广泛应用于网络服务、微服务架构、云计算基础设施、命令行工具等后端开发领域。
	
	总结:
	讨论了**Go语言**作为**开源编程语言**的**核心特点**，包括**简洁、高效、强大的并发能力**（如**Goroutines和Channels**）、**快速编译**及**垃圾回收**。对话还涵盖了Go语言在**网络服务、微服务架构**和**云计算**等**后端开发**中的主要**应用场景**。
	---

	请根据以下输入 Q&A 对，生成主题总结：
	`

	KnowLedgeExtractionSystemPrompt = `
	你是一个知识提取专家，你需要根据输入的Quesion-Answer对，提取出其中关于用户的**兴趣爱好**、**个人偏好**、**重要事件**等信息，并进行总结。

	以下是一些示例：
	单个Q&A对的情况：
	---
	Question: 我喜欢吃苹果
	Answer: 真好呀，苹果确实好吃
	
	你的输出:
	用户喜欢吃苹果
	---
	多个Q&A对的情况：
	---
	Question: 什么是Go语言？
	Answer: Go语言（或Golang）是Google开发的一种静态强类型、编译型、并发型，并具有垃圾回收功能的编程语言。
	
	Question: Go语言的主要特点是什么？
	Answer: 主要特点包括：静态类型、编译型语言、内置垃圾回收、原生支持并发（Goroutines和Channels）、以及快速的编译速度。
	
	你的输出:
	用户正在学习Go语言，并了解其主要特点如静态类型、编译型、垃圾回收和并发支持。
	---
	`
)

func newSegmentOverviewTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage(SegmentOverviewSystemPrompt),
		schema.UserMessage("{{.qas}}"),
	)
}

func newKnowledgeExtractionTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage(KnowLedgeExtractionSystemPrompt),
		schema.UserMessage("{{.qas}}"),
	)
}
