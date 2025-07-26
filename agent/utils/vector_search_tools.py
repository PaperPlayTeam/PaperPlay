from langchain.tools import tool
from .vector_store_manager import VectorStoreManager
import json
import logging
from typing import Dict, Any

# 创建全局的向量管理器实例
_vector_manager = None

def get_vector_manager() -> VectorStoreManager:
    """获取向量管理器实例（单例模式）"""
    global _vector_manager
    if _vector_manager is None:
        _vector_manager = VectorStoreManager()
    return _vector_manager

@tool
def search_similar_papers_tool(query_text: str, max_results: int = 5) -> str:
    """
    搜索与查询文本相似的论文
    
    Args:
        query_text (str): 查询文本，可以是问题、关键词或描述
        max_results (int): 返回的最大结果数量，默认5个
        
    Returns:
        str: 包含相似论文信息的JSON字符串，包括标题、作者、相似度等
    """
    try:
        vector_manager = get_vector_manager()
        
        # 限制最大结果数量
        n_results = min(max_results, 10)
        
        results = vector_manager.search_similar_papers(query_text, n_results)
        
        if not results:
            return "未找到相似的论文。可能向量库中还没有论文数据，或者查询文本没有匹配的内容。"
        
        # 格式化结果
        formatted_results = []
        for i, result in enumerate(results, 1):
            metadata = result.get('metadata', {})
            formatted_result = {
                "排名": i,
                "相似度": f"{result.get('similarity', 0):.3f}",
                "论文标题": metadata.get('title', '未知'),
                "作者": metadata.get('authors', '未知'),
                "年份": metadata.get('year', '未知'),
                "arXiv ID": metadata.get('arxiv_id', '未知'),
                "文档摘要": result.get('document', '')[:200] + "..." if len(result.get('document', '')) > 200 else result.get('document', ''),
                "向量ID": result.get('vector_id', '')
            }
            formatted_results.append(formatted_result)
        
        return f"""
找到了 {len(results)} 篇相似论文：

{json.dumps(formatted_results, ensure_ascii=False, indent=2)}

搜索完成！以上论文按相似度从高到低排序。
"""
        
    except Exception as e:
        return f"搜索相似论文时发生错误: {str(e)}"

@tool
def search_similar_concepts_tool(query_text: str, max_results: int = 5) -> str:
    """
    搜索与查询文本相似的概念
    
    Args:
        query_text (str): 查询文本，可以是概念名称、描述或问题
        max_results (int): 返回的最大结果数量，默认5个
        
    Returns:
        str: 包含相似概念信息的JSON字符串，包括概念名称、描述、相似度等
    """
    try:
        vector_manager = get_vector_manager()
        
        # 限制最大结果数量
        n_results = min(max_results, 10)
        
        results = vector_manager.search_similar_concepts(query_text, n_results)
        
        if not results:
            return "未找到相似的概念。可能向量库中还没有概念数据，或者查询文本没有匹配的内容。"
        
        # 格式化结果
        formatted_results = []
        for i, result in enumerate(results, 1):
            metadata = result.get('metadata', {})
            formatted_result = {
                "排名": i,
                "相似度": f"{result.get('similarity', 0):.3f}",
                "概念名称": metadata.get('concept_name', '未知'),
                "类别": metadata.get('category', '未知'),
                "难度级别": metadata.get('difficulty_level', '未知'),
                "概念描述": result.get('document', '')[:300] + "..." if len(result.get('document', '')) > 300 else result.get('document', ''),
                "向量ID": result.get('vector_id', '')
            }
            formatted_results.append(formatted_result)
        
        return f"""
找到了 {len(results)} 个相似概念：

{json.dumps(formatted_results, ensure_ascii=False, indent=2)}

搜索完成！以上概念按相似度从高到低排序。
"""
        
    except Exception as e:
        return f"搜索相似概念时发生错误: {str(e)}"

@tool
def hybrid_search_tool(query_text: str, search_papers: bool = True, search_concepts: bool = True, max_results: int = 5) -> str:
    """
    混合搜索论文和概念
    
    Args:
        query_text (str): 查询文本
        search_papers (bool): 是否搜索论文，默认True
        search_concepts (bool): 是否搜索概念，默认True
        max_results (int): 每种类型的最大结果数量，默认5个
        
    Returns:
        str: 包含论文和概念搜索结果的综合信息
    """
    try:
        vector_manager = get_vector_manager()
        
        # 限制最大结果数量
        n_results = min(max_results, 8)
        
        results = vector_manager.hybrid_search(
            query_text, 
            search_papers=search_papers, 
            search_concepts=search_concepts, 
            n_results=n_results
        )
        
        response_parts = []
        
        if search_papers and results.get("papers"):
            paper_results = []
            for i, result in enumerate(results["papers"], 1):
                metadata = result.get('metadata', {})
                paper_result = {
                    "排名": i,
                    "相似度": f"{result.get('similarity', 0):.3f}",
                    "标题": metadata.get('title', '未知'),
                    "作者": metadata.get('authors', '未知'),
                    "年份": metadata.get('year', '未知')
                }
                paper_results.append(paper_result)
            
            response_parts.append(f"""
📄 相关论文 ({len(paper_results)} 篇):
{json.dumps(paper_results, ensure_ascii=False, indent=2)}
""")
        
        if search_concepts and results.get("concepts"):
            concept_results = []
            for i, result in enumerate(results["concepts"], 1):
                metadata = result.get('metadata', {})
                concept_result = {
                    "排名": i,
                    "相似度": f"{result.get('similarity', 0):.3f}",
                    "概念": metadata.get('concept_name', '未知'),
                    "类别": metadata.get('category', '未知'),
                    "难度": metadata.get('difficulty_level', '未知')
                }
                concept_results.append(concept_result)
            
            response_parts.append(f"""
🧠 相关概念 ({len(concept_results)} 个):
{json.dumps(concept_results, ensure_ascii=False, indent=2)}
""")
        
        if not response_parts:
            return "未找到相关的论文或概念。"
        
        return "🔍 混合搜索结果：\n" + "\n".join(response_parts)
        
    except Exception as e:
        return f"混合搜索时发生错误: {str(e)}"

@tool
def get_vector_store_stats_tool() -> str:
    """
    获取向量库的统计信息
    
    Returns:
        str: 向量库统计信息，包括论文数量、概念数量、嵌入模型等
    """
    try:
        vector_manager = get_vector_manager()
        stats = vector_manager.get_collection_stats()
        
        if not stats:
            return "无法获取向量库统计信息。"
        
        formatted_stats = {
            "论文向量数量": stats.get('papers_count', 0),
            "概念向量数量": stats.get('concepts_count', 0),
            "总向量数量": stats.get('total_vectors', 0),
            "嵌入模型": stats.get('embedding_model', '未知'),
            "状态": "正常运行" if stats.get('total_vectors', 0) > 0 else "向量库为空"
        }
        
        return f"""
📊 向量库统计信息：

{json.dumps(formatted_stats, ensure_ascii=False, indent=2)}

向量库运行正常，可以进行相似性搜索。
"""
        
    except Exception as e:
        return f"获取向量库统计信息时发生错误: {str(e)}"

@tool
def add_paper_to_vector_store_tool(paper_text: str, paper_id: int, title: str = "", authors: str = "", arxiv_id: str = "") -> str:
    """
    将论文文本添加到向量库中
    
    Args:
        paper_text (str): 论文的文本内容（通常是简化后的文本）
        paper_id (int): 论文在数据库中的ID
        title (str): 论文标题，可选
        authors (str): 论文作者，可选  
        arxiv_id (str): arXiv ID，可选
        
    Returns:
        str: 添加结果信息
    """
    try:
        vector_manager = get_vector_manager()
        
        # 准备元数据
        metadata = {
            "title": title or "未知标题",
            "authors": authors or "未知作者",
            "arxiv_id": arxiv_id or "未知"
        }
        
        # 添加向量
        vector_id = vector_manager.add_paper_embedding(
            paper_id=paper_id,
            simplified_text=paper_text,
            metadata=metadata
        )
        
        return f"""
✅ 成功将论文添加到向量库！

- 向量ID: {vector_id}
- 论文ID: {paper_id}
- 标题: {title or '未提供'}
- 文本长度: {len(paper_text)} 字符

现在可以通过相似性搜索找到这篇论文了。
"""
        
    except Exception as e:
        return f"添加论文到向量库失败: {str(e)}" 