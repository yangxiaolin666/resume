package com.resume.company.service;

import java.util.List;
import java.util.Map;

public interface FavoriteService {

    /**
     * 批量检查内容是否被用户收藏
     * @param userId 用户ID
     * @param contentIds 内容ID列表
     * @return Map<内容ID, 是否收藏>
     */

    /**
     * 添加收藏
     * @param userId 用户ID
     * @param contentId 内容ID
     */
    void addFavorite(Integer userId, Integer contentId);

    Map<Integer, Boolean> batchIsFavorite(Integer userId, List<Integer> contentIds);

    /**
     * 取消收藏
     * @param userId 用户ID
     * @param contentId 内容ID
     */
    void removeFavorite(Integer userId, Integer contentId);

}
