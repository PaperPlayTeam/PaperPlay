import sqlite3
import json
import os
import logging
from typing import List, Dict, Any, Optional, Tuple
from pathlib import Path

class ConceptDatabaseManager:
    """概念论文数据库管理类 - 专门处理论文概念提取相关的数据"""
    
    def __init__(self, db_path: str = "paper_data/concept_papers.db"):
        self.db_path = db_path
        self.schema_path = "paper_data/concept_papers_schema.sql"
        self.logger = logging.getLogger(__name__)
        self._init_database()
    
    def _init_database(self):
        """初始化数据库，创建表结构"""
        try:
            # 确保目录存在
            os.makedirs(os.path.dirname(self.db_path), exist_ok=True)
            
            # 执行schema文件
            if os.path.exists(self.schema_path):
                with open(self.schema_path, 'r', encoding='utf-8') as f:
                    schema_sql = f.read()
                
                with sqlite3.connect(self.db_path) as conn:
                    conn.executescript(schema_sql)
                    conn.commit()
                self.logger.info(f"概念数据库初始化完成: {self.db_path}")
            else:
                self.logger.error(f"Schema文件不存在: {self.schema_path}")
        except Exception as e:
            self.logger.error(f"概念数据库初始化失败: {e}")
            raise
    
    def get_connection(self) -> sqlite3.Connection:
        """获取数据库连接"""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row  # 返回字典式的行对象
        return conn
    
    # ========== 论文相关操作 ==========
    
    def insert_paper(self, arxiv_id: str, title: str, authors: List[str], 
                    abstract: str, full_text: str, year: int = None,
                    citation_count: int = 0, journal: str = "", doi: str = "") -> int:
        """插入论文记录"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("""
                    INSERT INTO papers (arxiv_id, title, authors, abstract, full_text, 
                                      year, citation_count, journal, doi)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (arxiv_id, title, json.dumps(authors, ensure_ascii=False), 
                     abstract, full_text, year, citation_count, journal, doi))
                
                paper_id = cursor.lastrowid
                conn.commit()
                self.logger.info(f"成功插入论文: {title}, ID: {paper_id}")
                return paper_id
        except sqlite3.IntegrityError:
            # 如果arXiv ID已存在，返回现有记录的ID
            existing_paper = self.get_paper_by_arxiv_id(arxiv_id)
            if existing_paper:
                self.logger.info(f"论文已存在: {arxiv_id}, ID: {existing_paper['id']}")
                return existing_paper['id']
            raise
        except Exception as e:
            self.logger.error(f"插入论文失败: {e}")
            raise
    
    def get_paper_by_id(self, paper_id: int) -> Optional[Dict]:
        """根据ID查询论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers WHERE id = ?", (paper_id,))
                row = cursor.fetchone()
                if row:
                    result = dict(row)
                    result['authors'] = json.loads(result['authors'])
                    return result
                return None
        except Exception as e:
            self.logger.error(f"查询论文失败: {e}")
            return None
    
    def get_paper_by_arxiv_id(self, arxiv_id: str) -> Optional[Dict]:
        """根据arXiv ID查询论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers WHERE arxiv_id = ?", (arxiv_id,))
                row = cursor.fetchone()
                if row:
                    result = dict(row)
                    result['authors'] = json.loads(result['authors'])
                    return result
                return None
        except Exception as e:
            self.logger.error(f"查询论文失败: {e}")
            return None
    
    def get_all_papers(self) -> List[Dict]:
        """获取所有论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers ORDER BY created_at DESC")
                results = []
                for row in cursor.fetchall():
                    result = dict(row)
                    result['authors'] = json.loads(result['authors'])
                    results.append(result)
                return results
        except Exception as e:
            self.logger.error(f"查询论文列表失败: {e}")
            return []
    
    # ========== 概念相关操作 ==========
    
    def insert_concepts(self, paper_id: int, concepts: List[Dict]) -> List[int]:
        """批量插入概念"""
        concept_ids = []
        try:
            with self.get_connection() as conn:
                for i, concept in enumerate(concepts, 1):
                    cursor = conn.execute("""
                        INSERT INTO concepts (paper_id, concept_name, concept_explanation, 
                                            concept_order, importance_score)
                        VALUES (?, ?, ?, ?, ?)
                    """, (paper_id, concept['name'], concept['explanation'], 
                         i, concept.get('importance_score', 0.8)))
                    
                    concept_ids.append(cursor.lastrowid)
                
                conn.commit()
                self.logger.info(f"成功插入 {len(concepts)} 个概念，论文ID: {paper_id}")
                return concept_ids
        except Exception as e:
            self.logger.error(f"插入概念失败: {e}")
            raise
    
    def get_concepts_by_paper_id(self, paper_id: int) -> List[Dict]:
        """获取论文的所有概念"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("""
                    SELECT * FROM concepts 
                    WHERE paper_id = ? 
                    ORDER BY concept_order
                """, (paper_id,))
                return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            self.logger.error(f"查询概念失败: {e}")
            return []
    
    def get_concept_by_id(self, concept_id: int) -> Optional[Dict]:
        """根据ID查询概念"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM concepts WHERE id = ?", (concept_id,))
                row = cursor.fetchone()
                return dict(row) if row else None
        except Exception as e:
            self.logger.error(f"查询概念失败: {e}")
            return None
    
    def search_concepts_by_name(self, keyword: str) -> List[Dict]:
        """根据概念名称搜索"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("""
                    SELECT c.*, p.title as paper_title, p.arxiv_id 
                    FROM concepts c
                    JOIN papers p ON c.paper_id = p.id
                    WHERE c.concept_name LIKE ? 
                    ORDER BY c.importance_score DESC
                """, (f"%{keyword}%",))
                return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            self.logger.error(f"搜索概念失败: {e}")
            return []
    
    # ========== 完整论文信息查询 ==========
    
    def get_paper_with_concepts(self, paper_id: int) -> Optional[Dict]:
        """获取论文及其所有概念"""
        paper = self.get_paper_by_id(paper_id)
        if not paper:
            return None
        
        concepts = self.get_concepts_by_paper_id(paper_id)
        paper['concepts'] = concepts
        return paper
    
    def get_paper_with_concepts_by_arxiv_id(self, arxiv_id: str) -> Optional[Dict]:
        """根据arXiv ID获取论文及其所有概念"""
        paper = self.get_paper_by_arxiv_id(arxiv_id)
        if not paper:
            return None
        
        concepts = self.get_concepts_by_paper_id(paper['id'])
        paper['concepts'] = concepts
        return paper
    
    # ========== 统计相关操作 ==========
    
    def get_database_stats(self) -> Dict[str, Any]:
        """获取数据库统计信息"""
        try:
            with self.get_connection() as conn:
                stats = {}
                
                # 论文统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM papers")
                stats['total_papers'] = cursor.fetchone()['count']
                
                # 概念统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM concepts")
                stats['total_concepts'] = cursor.fetchone()['count']
                
                # 平均每篇论文的概念数
                cursor = conn.execute("""
                    SELECT AVG(concept_count) as avg_concepts
                    FROM (
                        SELECT COUNT(*) as concept_count 
                        FROM concepts 
                        GROUP BY paper_id
                    )
                """)
                avg_result = cursor.fetchone()
                stats['avg_concepts_per_paper'] = round(avg_result['avg_concepts'] or 0, 2)
                
                # 年份分布
                cursor = conn.execute("""
                    SELECT year, COUNT(*) as count 
                    FROM papers 
                    WHERE year IS NOT NULL
                    GROUP BY year 
                    ORDER BY year DESC
                """)
                stats['papers_by_year'] = {row['year']: row['count'] for row in cursor.fetchall()}
                
                # 最高引用的论文
                cursor = conn.execute("""
                    SELECT title, citation_count, arxiv_id 
                    FROM papers 
                    ORDER BY citation_count DESC 
                    LIMIT 5
                """)
                stats['top_cited_papers'] = [dict(row) for row in cursor.fetchall()]
                
                return stats
        except Exception as e:
            self.logger.error(f"获取统计信息失败: {e}")
            return {}
    
    # ========== 数据管理操作 ==========
    
    def delete_paper(self, paper_id: int) -> bool:
        """删除论文（级联删除概念）"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("DELETE FROM papers WHERE id = ?", (paper_id,))
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"删除论文成功: {paper_id}")
                    return True
                else:
                    self.logger.warning(f"论文不存在: {paper_id}")
                    return False
        except Exception as e:
            self.logger.error(f"删除论文失败: {e}")
            return False
    
    def update_paper_concepts(self, paper_id: int, concepts: List[Dict]) -> bool:
        """更新论文的概念（删除旧概念，插入新概念）"""
        try:
            with self.get_connection() as conn:
                # 删除旧概念
                conn.execute("DELETE FROM concepts WHERE paper_id = ?", (paper_id,))
                
                # 插入新概念
                for i, concept in enumerate(concepts, 1):
                    conn.execute("""
                        INSERT INTO concepts (paper_id, concept_name, concept_explanation, 
                                            concept_order, importance_score)
                        VALUES (?, ?, ?, ?, ?)
                    """, (paper_id, concept['name'], concept['explanation'], 
                         i, concept.get('importance_score', 0.8)))
                
                conn.commit()
                self.logger.info(f"更新论文概念成功: {paper_id}, 概念数: {len(concepts)}")
                return True
        except Exception as e:
            self.logger.error(f"更新论文概念失败: {e}")
            return False
    
    def close(self):
        """清理资源"""
        pass 