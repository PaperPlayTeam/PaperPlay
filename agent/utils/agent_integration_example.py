#!/usr/bin/env python3
"""
Agent集成示例 - 展示如何在LangChain Agent中使用向量搜索工具
"""

import sys
import os

# 添加项目根目录到Python路径
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

from langchain_community.chat_models import ChatTongyi
from langgraph.checkpoint.memory import MemorySaver
from langgraph.prebuilt import create_react_agent
from utils import (
    pdf_text_extraction_tool,
    search_similar_papers_tool,
    search_similar_concepts_tool,
    hybrid_search_tool,
    get_vector_store_stats_tool,
    add_paper_to_vector_store_tool
)

# 加载.env文件中的环境变量
try:
    from dotenv import load_dotenv
    load_dotenv()
    print("✅ 已加载.env文件中的环境变量")
except ImportError:
    print("⚠️  未安装python-dotenv包，无法加载.env文件")

def create_enhanced_paper_agent():
    """创建增强的论文处理代理，包含向量搜索功能"""
    
    # 检查API密钥
    api_key = os.getenv("DASHSCOPE_API_KEY")
    if not api_key:
        print("❌ 错误: 未找到DASHSCOPE_API_KEY环境变量")
        print("   请在.env文件中添加: DASHSCOPE_API_KEY=your-api-key")
        return None
    
    # 初始化通义千问模型
    model = ChatTongyi(
        model="qwen-max-latest",
        top_p=0.8,
        streaming=True,
        temperature=0.7,
        max_retries=3
    )
    
    # 创建工具列表 - 包含PDF处理和向量搜索功能
    tools = [
        pdf_text_extraction_tool,           # PDF文本提取
        search_similar_papers_tool,         # 搜索相似论文
        search_similar_concepts_tool,       # 搜索相似概念
        hybrid_search_tool,                 # 混合搜索
        get_vector_store_stats_tool,        # 向量库统计
        add_paper_to_vector_store_tool      # 添加论文到向量库
    ]
    
    # 创建memory
    memory = MemorySaver()
    
    # 创建ReAct代理
    agent_executor = create_react_agent(model, tools, checkpointer=memory)
    
    return agent_executor

def run_agent_with_message(agent_executor, message: str, thread_id: str = "demo"):
    """运行代理并处理消息"""
    config = {"configurable": {"thread_id": thread_id}}
    
    print(f"\n🤖 用户: {message}")
    print("🤖 助手:")
    
    try:
        for chunk in agent_executor.stream(
            {"messages": [("user", message)]}, config
        ):
            if "agent" in chunk:
                print(chunk["agent"]["messages"][0].content)
            elif "tools" in chunk:
                # 可选：显示工具调用信息
                pass
    except Exception as e:
        print(f"❌ 处理消息时出错: {e}")

def demo_agent_capabilities():
    """演示代理的各种功能"""
    
    # 创建代理
    agent_executor = create_enhanced_paper_agent()
    if not agent_executor:
        return
    
    print("🚀 增强版论文处理代理演示")
    print("=" * 60)
    
    # 示例对话场景
    demo_conversations = [
        # 1. 检查向量库状态
        "请检查一下向量库的状态，告诉我现在有多少论文和概念。",
        
        # 2. PDF处理请求
        "请帮我提取 /home/bugsmith/paperplay/papers/1706.03762v7.pdf 中的文本信息，然后将论文内容添加到向量库中。",
        
        # 3. 向量搜索请求
        "我想了解attention机制相关的论文，请帮我搜索5篇最相关的论文。",
        
        # 4. 混合搜索请求
        "关于Transformer架构，请同时搜索相关的论文和概念，我想全面了解这个主题。",
        
        # 5. 概念搜索请求
        "请搜索与'深度学习'相关的概念，帮我理解这个领域的核心概念。"
    ]
    
    for i, message in enumerate(demo_conversations, 1):
        print(f"\n{'='*20} 对话 {i} {'='*20}")
        run_agent_with_message(agent_executor, message, thread_id=f"demo_{i}")
        
        # 添加分隔符
        print("\n" + "-" * 60)
    
    print("\n🎉 代理功能演示完成！")
    
    # 交互式对话
    print("\n💬 现在你可以与代理进行交互对话（输入'quit'退出）:")
    thread_id = "interactive"
    
    while True:
        try:
            user_input = input("\n👤 你: ").strip()
            if user_input.lower() in ['quit', 'exit', '退出', 'q']:
                print("👋 再见！")
                break
            
            if user_input:
                run_agent_with_message(agent_executor, user_input, thread_id)
        
        except KeyboardInterrupt:
            print("\n👋 再见！")
            break
        except Exception as e:
            print(f"❌ 发生错误: {e}")

def main():
    """主函数"""
    print("📖 Agent集成示例")
    print("这个示例展示了如何将PDF处理和向量搜索工具集成到LangChain Agent中")
    print()
    
    # 检查环境
    print("🔧 环境检查:")
    print(f"  - Python路径: {sys.executable}")
    print(f"  - 工作目录: {os.getcwd()}")
    print(f"  - DASHSCOPE_API_KEY: {'✅ 已设置' if os.getenv('DASHSCOPE_API_KEY') else '❌ 未设置'}")
    print()
    
    # 运行演示
    demo_agent_capabilities()

if __name__ == "__main__":
    main() 