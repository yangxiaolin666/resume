package com.resume.company.service.impl;

import com.resume.company.service.FavoriteService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class FavoriteServiceImpl implements FavoriteService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_FAVORITE = "user:favorite:";

    private String getKey(Integer userId) {
        return KEY_FAVORITE + userId;
    }

    @Override
    public Map<Integer, Boolean> batchIsFavorite(Integer userId, List<Integer> contentIds) {
        if (contentIds == null || contentIds.isEmpty()) {
            return new HashMap<>();
        }

        String key = getKey(userId);
        Map<Integer, Boolean> map = new HashMap<>();

        for (Integer contentId : contentIds) {
            Boolean isMember = redisTemplate.opsForSet().isMember(key, String.valueOf(contentId));
            map.put(contentId, Boolean.TRUE.equals(isMember));
        }

        return map;
    }



    @Override
    public void addFavorite(Integer userId, Integer contentId) {
        String key = getKey(userId);
        redisTemplate.opsForSet().add(key, String.valueOf(contentId));
    }

    @Override
    public void removeFavorite(Integer userId, Integer contentId) {
        String key = getKey(userId);
        redisTemplate.opsForSet().remove(key, String.valueOf(contentId));
    }

}
