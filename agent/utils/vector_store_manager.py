import chromadb
from chromadb.config import Settings
from sentence_transformers import SentenceTransformer
import os
import logging
from typing import List, Dict, Any, Optional
import uuid

class VectorStoreManager:
    """向量数据库管理类 - 处理Chroma向量数据库的所有操作"""
    
    def __init__(self, 
                 persist_directory: str = "chroma_db",
                 embedding_model_name: str = "sentence-transformers/all-MiniLM-L6-v2"):
        self.persist_directory = persist_directory
        self.embedding_model_name = embedding_model_name
        self.logger = logging.getLogger(__name__)
        
        # 确保目录存在
        os.makedirs(persist_directory, exist_ok=True)
        
        # 初始化Chroma客户端
        self.client = chromadb.PersistentClient(
            path=persist_directory,
            settings=Settings(
                anonymized_telemetry=False,
                allow_reset=True
            )
        )
        
        # 初始化嵌入模型
        self.embedding_model = SentenceTransformer(embedding_model_name)
        
        # 创建集合
        self.papers_collection = self._get_or_create_collection("papers")
        self.concepts_collection = self._get_or_create_collection("concepts")
        
        self.logger.info("向量数据库管理器初始化完成")
    
    def _get_or_create_collection(self, collection_name: str):
        """获取或创建集合"""
        try:
            collection = self.client.get_collection(collection_name)
            self.logger.info(f"获取现有集合: {collection_name}")
        except:
            collection = self.client.create_collection(
                name=collection_name,
                metadata={"hnsw:space": "cosine"}
            )
            self.logger.info(f"创建新集合: {collection_name}")
        return collection
    
    def generate_embedding(self, text: str) -> List[float]:
        """生成文本的嵌入向量"""
        try:
            embedding = self.embedding_model.encode(text).tolist()
            return embedding
        except Exception as e:
            self.logger.error(f"生成嵌入向量失败: {e}")
            raise
    
    # ========== 论文向量操作 ==========
    
    def add_paper_embedding(self, paper_id: int, simplified_text: str, 
                           metadata: Dict[str, Any] = None) -> str:
        """添加论文的嵌入向量"""
        try:
            # 生成唯一的向量ID  
            vector_id = f"paper_{paper_id}_{uuid.uuid4().hex[:8]}"
            
            # 生成嵌入向量
            embedding = self.generate_embedding(simplified_text)
            
            # 准备元数据
            paper_metadata = {
                "paper_id": paper_id,
                "source_type": "paper",
                "text_length": len(simplified_text)
            }
            if metadata:
                paper_metadata.update(metadata)
            
            # 添加到集合
            self.papers_collection.add(
                embeddings=[embedding],
                documents=[simplified_text],
                metadatas=[paper_metadata],
                ids=[vector_id]
            )
            
            self.logger.info(f"成功添加论文向量: {vector_id}")
            return vector_id
            
        except Exception as e:
            self.logger.error(f"添加论文向量失败: {e}")
            raise
    
    def search_similar_papers(self, query_text: str, n_results: int = 5) -> List[Dict]:
        """搜索相似的论文"""
        try:
            # 生成查询向量
            query_embedding = self.generate_embedding(query_text)
            
            # 执行搜索
            results = self.papers_collection.query(
                query_embeddings=[query_embedding],
                n_results=n_results,
                include=['documents', 'metadatas', 'distances']
            )
            
            # 格式化结果
            formatted_results = []
            for i in range(len(results['ids'][0])):
                formatted_results.append({
                    'vector_id': results['ids'][0][i],
                    'document': results['documents'][0][i],
                    'metadata': results['metadatas'][0][i],
                    'distance': results['distances'][0][i],
                    'similarity': 1 - results['distances'][0][i]  # 转换为相似度
                })
            
            self.logger.info(f"搜索到{len(formatted_results)}个相似论文")
            return formatted_results
            
        except Exception as e:
            self.logger.error(f"搜索相似论文失败: {e}")
            return []
    
    def get_paper_embedding(self, vector_id: str) -> Optional[Dict]:
        """获取指定论文的向量信息"""
        try:
            results = self.papers_collection.get(
                ids=[vector_id],
                include=['documents', 'metadatas', 'embeddings']
            )
            
            if results['ids']:
                return {
                    'vector_id': results['ids'][0],
                    'document': results['documents'][0],
                    'metadata': results['metadatas'][0],
                    'embedding': results['embeddings'][0]
                }
            return None
            
        except Exception as e:
            self.logger.error(f"获取论文向量失败: {e}")
            return None
    
    # ========== 概念向量操作 ==========
    
    def add_concept_embedding(self, concept_id: int, concept_text: str, 
                             metadata: Dict[str, Any] = None) -> str:
        """添加概念的嵌入向量"""
        try:
            # 生成唯一的向量ID
            vector_id = f"concept_{concept_id}_{uuid.uuid4().hex[:8]}"
            
            # 生成嵌入向量
            embedding = self.generate_embedding(concept_text)
            
            # 准备元数据
            concept_metadata = {
                "concept_id": concept_id,
                "source_type": "concept",
                "text_length": len(concept_text)
            }
            if metadata:
                concept_metadata.update(metadata)
            
            # 添加到集合
            self.concepts_collection.add(
                embeddings=[embedding],
                documents=[concept_text],
                metadatas=[concept_metadata],
                ids=[vector_id]
            )
            
            self.logger.info(f"成功添加概念向量: {vector_id}")
            return vector_id
            
        except Exception as e:
            self.logger.error(f"添加概念向量失败: {e}")
            raise
    
    def search_similar_concepts(self, query_text: str, n_results: int = 5) -> List[Dict]:
        """搜索相似的概念"""
        try:
            # 生成查询向量
            query_embedding = self.generate_embedding(query_text)
            
            # 执行搜索
            results = self.concepts_collection.query(
                query_embeddings=[query_embedding],
                n_results=n_results,
                include=['documents', 'metadatas', 'distances']
            )
            
            # 格式化结果
            formatted_results = []
            for i in range(len(results['ids'][0])):
                formatted_results.append({
                    'vector_id': results['ids'][0][i],
                    'document': results['documents'][0][i],
                    'metadata': results['metadatas'][0][i],
                    'distance': results['distances'][0][i],
                    'similarity': 1 - results['distances'][0][i]
                })
            
            self.logger.info(f"搜索到{len(formatted_results)}个相似概念")
            return formatted_results
            
        except Exception as e:
            self.logger.error(f"搜索相似概念失败: {e}")
            return []
    
    # ========== 混合搜索 ==========
    
    def hybrid_search(self, query_text: str, search_papers: bool = True, 
                     search_concepts: bool = True, n_results: int = 10) -> Dict[str, List[Dict]]:
        """混合搜索论文和概念"""
        results = {"papers": [], "concepts": []}
        
        if search_papers:
            results["papers"] = self.search_similar_papers(query_text, n_results)
        
        if search_concepts:
            results["concepts"] = self.search_similar_concepts(query_text, n_results)
        
        return results
    
    # ========== 批量操作 ==========
    
    def batch_add_paper_embeddings(self, papers_data: List[Dict]) -> List[str]:
        """批量添加论文向量"""
        vector_ids = []
        try:
            embeddings = []
            documents = []
            metadatas = []
            ids = []
            
            for paper_data in papers_data:
                paper_id = paper_data['paper_id']
                simplified_text = paper_data['simplified_text']
                metadata = paper_data.get('metadata', {})
                
                # 生成向量ID和嵌入
                vector_id = f"paper_{paper_id}_{uuid.uuid4().hex[:8]}"
                embedding = self.generate_embedding(simplified_text)
                
                # 准备数据
                paper_metadata = {
                    "paper_id": paper_id,
                    "source_type": "paper",
                    "text_length": len(simplified_text)
                }
                paper_metadata.update(metadata)
                
                ids.append(vector_id)
                embeddings.append(embedding)
                documents.append(simplified_text)
                metadatas.append(paper_metadata)
                vector_ids.append(vector_id)
            
            # 批量添加
            self.papers_collection.add(
                embeddings=embeddings,
                documents=documents,
                metadatas=metadatas,
                ids=ids
            )
            
            self.logger.info(f"成功批量添加{len(vector_ids)}个论文向量")
            return vector_ids
            
        except Exception as e:
            self.logger.error(f"批量添加论文向量失败: {e}")
            raise
    
    # ========== 统计和管理 ==========
    
    def get_collection_stats(self) -> Dict[str, Any]:
        """获取集合统计信息"""
        try:
            papers_count = self.papers_collection.count()
            concepts_count = self.concepts_collection.count()
            
            return {
                "papers_count": papers_count,
                "concepts_count": concepts_count,
                "total_vectors": papers_count + concepts_count,
                "embedding_model": self.embedding_model_name
            }
        except Exception as e:
            self.logger.error(f"获取统计信息失败: {e}")
            return {}
    
    def delete_paper_embeddings(self, paper_id: int):
        """删除指定论文的所有向量"""
        try:
            # 查找相关向量
            results = self.papers_collection.get(
                where={"paper_id": paper_id}
            )
            
            if results['ids']:
                self.papers_collection.delete(ids=results['ids'])
                self.logger.info(f"删除论文{paper_id}的{len(results['ids'])}个向量")
        except Exception as e:
            self.logger.error(f"删除论文向量失败: {e}")
    
    def reset_collections(self):
        """重置所有集合（谨慎使用）"""
        try:
            self.client.reset()
            self.papers_collection = self._get_or_create_collection("papers")
            self.concepts_collection = self._get_or_create_collection("concepts")
            self.logger.warning("已重置所有向量集合")
        except Exception as e:
            self.logger.error(f"重置集合失败: {e}") 