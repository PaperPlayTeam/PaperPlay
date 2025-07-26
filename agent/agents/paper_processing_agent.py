# Import relevant functionality
from langchain_community.chat_models import ChatTongyi
from langgraph.checkpoint.memory import MemorySaver
from langgraph.prebuilt import create_react_agent
from langchain.tools import tool
import sys
import os
import json

# 加载.env文件中的环境变量
try:
    from dotenv import load_dotenv
    load_dotenv()
    print("✅ 已加载.env文件中的环境变量")
except ImportError:
    print("⚠️  未安装python-dotenv包，无法加载.env文件")
    print("   可以通过 'pip install python-dotenv' 安装")

# 添加utils目录到Python路径
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

from utils import (
    PDFTextExtractor,
    DatabaseManager,
    VectorStoreManager,
    search_similar_papers_tool,
    search_similar_concepts_tool,
    hybrid_search_tool,
    get_vector_store_stats_tool,
    add_paper_to_vector_store_tool
)

# 创建PDF文本提取工具包装器
@tool
def pdf_text_extraction_tool(file_path: str) -> str:
    """
    从PDF文件中提取文本和元数据
    
    Args:
        file_path (str): PDF文件的路径
        
    Returns:
        str: 包含提取结果的JSON字符串，包括论文标题、作者、摘要、全文等信息
    """
    try:
        extractor = PDFTextExtractor()
        result = extractor.extract_text_from_pdf(file_path)
        
        if result["status"] == "success":
            paper = result["paper"]
            return f"""
PDF文本提取成功！

论文信息：
- 标题: {paper['title']}
- 作者: {', '.join(paper['authors']) if isinstance(paper['authors'], list) else paper['authors']}
- 摘要: {paper['abstract'][:300]}...
- ArXiv ID: {paper['arxiv_id']}
- 年份: {paper['year']}
- 全文长度: {paper['parsed_text_length']} 字符

提取的全文内容已包含在结果中。
"""
        else:
            return f"PDF文本提取失败: {result['message']}"
            
    except Exception as e:
        return f"处理PDF时发生错误: {str(e)}"

@tool
def complete_paper_processing_tool(file_path: str) -> str:
    """
    完整的论文处理流程：提取文本 -> 存储到数据库 -> 添加到向量库
    
    Args:
        file_path (str): PDF文件路径
        
    Returns:
        str: 处理结果描述
    """
    try:
        # 1. 提取PDF文本和元数据
        pdf_extractor = PDFTextExtractor()
        result = pdf_extractor.extract_text_from_pdf(file_path)
        
        if result["status"] != "success":
            return f"❌ PDF提取失败: {result['message']}"
        
        paper_data = result["paper"]
        
        # 2. 存储到传统数据库 (适配新API)
        db_manager = DatabaseManager()
        
        # 适配新的字段格式
        paper_author = "; ".join(paper_data["authors"]) if isinstance(paper_data["authors"], list) else str(paper_data["authors"])
        paper_pub_ym = str(paper_data.get("year", "未知"))  # 从年份转换为发表年月
        
        paper_id = db_manager.insert_paper(
            title=paper_data["title"],
            paper_author=paper_author,
            paper_pub_ym=paper_pub_ym,
            paper_citation_count="0"  # MVP版本默认为0
        )
        
        # 3. 添加到向量库
        vector_manager = VectorStoreManager()
        
        # 使用摘要和标题作为向量存储的文本  
        vector_text = f"{paper_data['title']}\n\n{paper_data['abstract']}"
        if len(vector_text) < 100:  # 如果摘要太短，使用部分正文
            vector_text += f"\n\n{paper_data['parsed_text'][:1000]}"
        
        vector_id = vector_manager.add_paper_embedding(
            paper_id=paper_id,
            simplified_text=vector_text,
            metadata={
                "title": paper_data["title"],
                "authors": paper_author,
                "year": paper_data.get("year", "未知"),
                "arxiv_id": paper_data.get("arxiv_id", ""),
                "file_path": file_path
            }
        )
        
        return f"""✅ 论文处理完成！
📄 标题: {paper_data['title']}
👥 作者: {paper_author}
📅 年份: {paper_data.get('year', '未知')}
🆔 数据库ID: {paper_id}
🔍 向量ID: {vector_id}
📊 数据来源: {result.get('metadata_source', '未知')}"""
        
    except Exception as e:
        return f"❌ 处理过程中发生错误: {str(e)}"

@tool
def get_database_stats_tool() -> str:
    """
    获取数据库统计信息
    
    Returns:
        str: 数据库统计信息
    """
    try:
        db_manager = DatabaseManager()
        stats = db_manager.get_system_stats()
        
        formatted_stats = {
            "总学科数": stats.get('total_subjects', 0),
            "总论文数": stats.get('total_papers', 0),
            "总关卡数": stats.get('total_levels', 0), 
            "总题目数": stats.get('total_questions', 0),
            "平均每关卡题目数": stats.get('avg_questions_per_level', 0),
            "平均题目分值": stats.get('avg_question_score', 0),
            "数据库状态": "正常运行"
        }
        
        # 格式化输出
        result = "📊 **数据库统计信息**\n\n"
        for key, value in formatted_stats.items():
            result += f"• {key}: {value}\n"
        
        return result
        
    except Exception as e:
        return f"❌ 获取统计信息失败: {str(e)}"

# Create the agent with ChatTongyi (通义千问)
memory = MemorySaver()

# 检查API密钥是否设置
api_key = os.getenv("DASHSCOPE_API_KEY")
if not api_key:
    print("❌ 错误: 未找到DASHSCOPE_API_KEY环境变量")
    print("   请在.env文件中添加: DASHSCOPE_API_KEY=your-api-key")
    sys.exit(1)

# 初始化通义千问模型
model = ChatTongyi(
    model="qwen-max-latest",
    top_p=0.8,
    streaming=True,
    temperature=0.7,
    max_retries=3
)

# 创建完整的工具列表
tools = [
    pdf_text_extraction_tool,           # PDF文本提取
    complete_paper_processing_tool,     # 完整论文处理流程
    search_similar_papers_tool,         # 搜索相似论文
    search_similar_concepts_tool,       # 搜索相似概念
    hybrid_search_tool,                 # 混合搜索
    get_vector_store_stats_tool,        # 向量库统计
    get_database_stats_tool,            # 数据库统计
    add_paper_to_vector_store_tool      # 手动添加论文到向量库
]

agent_executor = create_react_agent(model, tools, checkpointer=memory)

def run_paper_processing_agent(message: str, thread_id: str = "1") -> dict:
    """
    运行代理并处理用户消息
    
    Args:
        message (str): 用户消息
        thread_id (str): 线程ID，用于维护会话状态
        
    Returns:
        dict: 包含处理结果的字典
    """
    config = {"configurable": {"thread_id": thread_id}}
    
    responses = []
    
    try:
        for chunk in agent_executor.stream(
            {"messages": [("user", message)]}, config
        ):
            if "agent" in chunk:
                response = chunk["agent"]["messages"][0].content
                print(response)
                responses.append(response)
            elif "tools" in chunk:
                # 可选：显示工具调用信息
                print(f"🔧 工具调用: {chunk}")
        
        return {
            "status": "success",
            "responses": responses,
            "message": "处理完成"
        }
        
    except Exception as e:
        error_msg = f"❌ 运行出错: {e}"
        print(error_msg)
        return {
            "status": "error",
            "message": str(e)
        }

def process_single_paper(file_path: str) -> dict:
    """
    处理单个论文文件的便捷方法
    
    Args:
        file_path (str): PDF文件路径
        
    Returns:
        dict: 处理结果
    """
    message = f"请使用完整论文处理流程处理这个PDF文件: {file_path}"
    return run_paper_processing_agent(message, thread_id=f"paper_{hash(file_path)}")

if __name__ == "__main__":
    print("🤖 增强版论文处理代理已启动！")
    print("支持的功能:")
    print("  📄 PDF文本提取")
    print("  💾 数据库存储")
    print("  🔍 向量搜索")
    print("  📊 统计信息")
    print("你可以让我帮你完整处理PDF文件。")
    print("示例: '请完整处理 /path/to/paper.pdf'")
    print("=" * 50)
    
    # 示例用法
    user_message = "请完整处理 /home/bugsmith/paperplay/papers/1706.03762v7.pdf"
    print(f"示例消息: {user_message}")
    
    try:
        result = run_paper_processing_agent(user_message)
        print(f"\n处理结果: {result['status']}")
    except Exception as e:
        print(f"❌ 运行出错: {e}")
        print("请检查DASHSCOPE_API_KEY是否正确设置")