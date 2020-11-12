package session

import (
	"errors"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// sessionsPrefix is used to prefix redis keys that represent user sessions.
const sessionsPrefix = "sessions:"

// cmdGetSession attempts to retrieve a session from redis. Before querying for the session, the
// expired sessions are removed. If the session is found, then the expiration time of the
// individual session as well as the set containing all of the user's sessions is reset.
var cmdGetSession = redis.NewScript(1, `
	redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
	local res = redis.call('ZSCORE', KEYS[1], ARGV[2])
	if res == nil then
		return nil
	end
	redis.call('ZREM', KEYS[1], ARGV[2])
	redis.call('ZADD' KEYS[1], ARGV[3], ARGV[2])
	redis.call('EXPIRE', KEYS[1], ARGV[4])
	return res
`)

// ErrSessionNotFound is used when attempting to access a session token that does not exist in the
// session store.
var ErrSessionNotFound = errors.New("session not found")

// Store provides methods for interacting with a session store.
type Store interface {
	Get(sess Session) (Session, error)
	Add(sess Session) error
	Remove(sess Session) error
	RemoveAll(sess Session) error
	RemoveOthers(sess Session) error
	Close() error
}

// StoreConfig represents configuration options for a redis session store.
type StoreConfig struct {
	Redis      string
	SessionTTL time.Duration
	UserTTL    time.Duration
}

type store struct {
	redis      *redis.Pool
	sessionTTL time.Duration
	userTTL    time.Duration
}

// NewStore creates a new redis session store.
func NewStore(cfg StoreConfig) (Store, error) {
	r := &redis.Pool{
		IdleTimeout: 3 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis)
		},
	}

	conn := r.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return nil, err
	}

	s := &store{
		redis:      r,
		sessionTTL: cfg.SessionTTL,
		userTTL:    cfg.UserTTL,
	}
	return s, nil
}

// Get retrieves a session from the store.
func (s *store) Get(sess Session) (Session, error) {
	conn := s.redis.Get()
	defer conn.Close()

	now := time.Now()

	res, err := cmdGetSession.Do(conn, sessionsPrefix+strconv.Itoa(sess.ID), now.Unix(), sess.Key, now.Add(s.sessionTTL).Unix(), s.userTTL.Seconds)
	if err != nil {
		return Session{}, err
	}
	if res == nil {
		return Session{}, ErrSessionNotFound
	}
	return sess, nil
}

// Add adds a new session to the store.
func (s *store) Add(sess Session) error {
	conn := s.redis.Get()
	defer conn.Close()

	_, err := conn.Do("ZADD", sessionsPrefix+strconv.Itoa(sess.ID), time.Now().Add(s.sessionTTL).Unix(), sess.Key)
	return err
}

// Remove removes a user session from the store.
func (s *store) Remove(sess Session) error {
	conn := s.redis.Get()
	defer conn.Close()

	_, err := conn.Do("ZREM", sessionsPrefix+strconv.Itoa(sess.ID), sess.Key)
	return err
}

// RemoveAll removes all sessions for the given user.
func (s *store) RemoveAll(sess Session) error {
	conn := s.redis.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", sessionsPrefix+strconv.Itoa(sess.ID))
	return err
}

// RemoveOthers removes all sessions for the given user except for the current session that is
// represented by the supplied token.
func (s *store) RemoveOthers(sess Session) error {
	conn := s.redis.Get()
	defer conn.Close()

	sessionsKey := sessionsPrefix + strconv.Itoa(sess.ID)

	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("DEL", sessionsKey); err != nil {
		return err
	}
	if err := conn.Send("ZADD", sessionsKey, time.Now().Add(s.sessionTTL).Unix(), sess.Key); err != nil {
		return err
	}
	if err := conn.Send("EXPIRE", sessionsKey, s.userTTL.Seconds); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}

// Close closes the underlying redis connection.
func (s *store) Close() error {
	return s.redis.Close()
}
