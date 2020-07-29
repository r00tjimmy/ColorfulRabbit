package ColorfulRabbit

/**
redis struct 版本
 */
import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "log"
  "time"
)

type XRedis struct {
  Rds       redis.Conn
  RdsPool   *redis.Pool
}

func NewXR(host, port, pwd string, db int) (*XRedis, error) {
  //Rds, err := redis.Dial("tcp", host + ":" + port, redis.DialPassword(pwd), redis.DialDatabase(db))
  Rds, err := redis.Dial("tcp", host + ":" + port, redis.DialPassword(pwd), redis.DialDatabase(db), redis.DialConnectTimeout(2 * time.Second))
  //Rds, err := redis.DialTimeout("tcp", host + ":" + port, redis.DialPassword(pwd), redis.DialDatabase(db))
  CheckError(err, "redis newconn err")
  return &XRedis{ Rds: Rds}, err
}

func NewXrPool(host, port, pwd string, db int) (*XRedis, error) {
  rdsPool := &redis.Pool{
    MaxIdle:          20,
    MaxActive:        7000,
    IdleTimeout:      60 * time.Second,
    Wait:             true,
    Dial: func() (redis.Conn, error) {
      conn, err := redis.Dial("tcp", host + ":" + port, redis.DialPassword(pwd),
        redis.DialDatabase(db), redis.DialConnectTimeout(3 * time.Second))
      if err != nil {
        CheckError(err, "redis conn pool err")
        return nil, err
      }
      return conn, err
    },
  }
  return &XRedis{RdsPool:  rdsPool}, nil
}

func (x *XRedis) Close() error {
  x.Rds.Close()
  return nil
}

func (x *XRedis) GetConn() redis.Conn {
  if x.Rds != nil {
    return x.Rds
  }
  return x.RdsPool.Get()
}

func (x *XRedis) GetKey(pattern string) ([]string, error) {
  //conn := x.Rds
  conn := x.GetConn()
  //defer x.Close()

  iter := 0
  var keys []string
  for {
    //arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
    arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern, "COUNT", 5000))
    if err != nil {
      return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
    }

    iter, _ = redis.Int(arr[0], nil)
    k, _ := redis.Strings(arr[1], nil)
    keys = append(keys, k...)

    if iter == 0 {
      break
    }

    if len(k) > 0 {
      break
    }
  }

  return keys, nil
}

func (x *XRedis) GetKeys(pattern string) ([]string, error) {
  conn := x.GetConn()
  //conn := x.Rds
  //defer x.Close()

  iter := 0
  var keys []string
  for {
    //arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
    arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern, "COUNT", 5000))
    if err != nil {
      return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
    }

    iter, _ = redis.Int(arr[0], nil)
    k, _ := redis.Strings(arr[1], nil)
    keys = append(keys, k...)

    if iter == 0 {
      break
    }

  }

  return keys, nil
}


func (x *XRedis) XKeyExist(pattern string) ([]string, error) {
  // scan判断key是否存在
  //conn := x.Rds
  conn := x.GetConn()
  //defer x.Close()

  iter := 0
  var keys []string
  for {
    //arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
    arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern, "COUNT", 5000))
    if err != nil {
      return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
    }

    iter, _ = redis.Int(arr[0], nil)
    log.Println("xKeyExist iter ------------------ ", iter)
    k, _ := redis.Strings(arr[1], nil)
    log.Println("xKeyExist k ------------------ ", k)
    //if len(k) > 0 {
    //  os.Exit(1)
    //}
    keys = append(keys, k...)

    if iter == 0 {
      break
    }
  }

  return keys, nil
}



func (x *XRedis) HGetAll(key string, field ...string) (map[string]interface{}, error) {
  //conn := x.Rds
  conn := x.GetConn()
  //defer x.Close()
  keys, err := redis.Values(conn.Do("HKEYS", key))
  CheckError(err, "redis hmget error")
  //return keys, err
  vals, err := redis.Values(conn.Do("HVALS", key))

  hmAll := make(map[string]interface{})
  for i, key := range keys {
    hmAll[string(key.([]uint8))] = string(vals[i].([]uint8))
  }
  return hmAll, nil
}


func (x *XRedis) Get(key string) ([]byte, error) {
  //conn := x.Rds
  conn := x.GetConn()
  //defer x.Close()
  var data []byte
  data, err := redis.Bytes(conn.Do("GET", key))
  if err != nil {
    return data, fmt.Errorf("error get key %s: %v", key, err)
  }
  return data, err
}

func (x *XRedis) Set(key string, val string) ([]byte, error) {
  //conn := x.Rds
  conn := x.GetConn()
  //defer x.Close()
  var data []byte
  data, err := redis.Bytes(conn.Do("SET", key, val))
  if err != nil {
    return data, fmt.Errorf("error set key %s: %v", key, err)
  }
  return data, err
}

func (x *XRedis) HSetAll(key string, m map[string]interface{}) error {
  //conn := x.Rds
  conn := x.GetConn()
  _, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(m)...)
  CheckError(err, "redis HSetAll error")
  return err
}




