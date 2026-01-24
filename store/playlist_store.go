package store

import (
	"errors"
	"time"

	"github.com/ProjectDistribute/distributor/model"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

var ErrPlaylistExists = errors.New("playlist already exists")

type PlaylistStore struct {
	db *gorm.DB
}

func NewPlaylistStore(db *gorm.DB) *PlaylistStore {
	return &PlaylistStore{db: db}
}

func (ps *PlaylistStore) GetUserPlaylists(userID uuid.UUID) ([]model.Playlist, error) {
	var playlists []model.Playlist
	if err := ps.db.Where("user_id = ?", userID).Find(&playlists).Error; err != nil {
		return nil, err
	}
	return playlists, nil
}

func (ps *PlaylistStore) CreatePlaylist(playlist *model.Playlist) (model.Playlist, error) {
	if err := ps.db.Create(playlist).Error; err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return model.Playlist{}, ErrPlaylistExists
			}
		}
		return model.Playlist{}, err
	}
	return *playlist, nil
}

func (ps *PlaylistStore) IsFolderOwnedByUser(folderID uuid.UUID, userID uuid.UUID) bool {
	var folder model.PlaylistFolder

	if err := ps.db.First(&folder, "id = ? AND user_id = ?", folderID, userID).Error; err != nil {
		return false
	}
	return folder.UserID == userID
}

func (ps *PlaylistStore) GetLibrary(userID uuid.UUID) ([]model.PlaylistFolder, []model.Playlist, error) {
	var dbFolders []model.PlaylistFolder
	var dbPlaylists []model.Playlist

	// 1. Fetch all data
	if err := ps.db.Where("user_id = ?", userID).Find(&dbFolders).Error; err != nil {
		return nil, nil, err
	}
	if err := ps.db.Where("user_id = ?", userID).Find(&dbPlaylists).Error; err != nil {
		return nil, nil, err
	}
	return dbFolders, dbPlaylists, nil
}

func (ps *PlaylistStore) CreatePlaylistFolder(folder *model.PlaylistFolder) (model.PlaylistFolder, error) {
	if err := ps.db.Create(folder).Error; err != nil {
		return model.PlaylistFolder{}, err
	}
	return *folder, nil
}

func (ps *PlaylistStore) DeletePlaylistFolder(folderID uuid.UUID, userID uuid.UUID, adminOverride bool) error {
	if adminOverride {
		return ps.db.Delete(&model.PlaylistFolder{}, "id = ?", folderID).Error
	}
	return ps.db.Delete(&model.PlaylistFolder{}, "id = ? AND user_id = ?", folderID, userID).Error
}

func (ps *PlaylistStore) GetPlaylistByID(playlistID uuid.UUID) (*model.Playlist, error) {
	var playlist model.Playlist
	if err := ps.db.
		Preload("PlaylistFolder.User").
		Preload("Songs").
		Preload("Songs.SongFiles").
		// Preload("Songs.Album").
		// Preload("Songs.Album.Artist").
		First(&playlist, "id = ?", playlistID).
		Error; err != nil {
		return nil, err
	}

	return &playlist, nil
}

func (ps *PlaylistStore) AddSongToPlaylist(playlistID uuid.UUID, songID uuid.UUID) error {
	playlist := model.Playlist{ID: playlistID}
	song := model.Song{ID: songID}

	err := ps.db.Model(&playlist).Association("Songs").Append(&song)
	if err != nil {
		return err
	}

	// Force update UpdatedAt
	ps.db.Model(&playlist).Update("updated_at", time.Now())

	return nil
}

func (ps *PlaylistStore) RemoveSongFromPlaylist(playlistID uuid.UUID, songID uuid.UUID) error {
	playlist := model.Playlist{ID: playlistID}
	song := model.Song{ID: songID}

	err := ps.db.Model(&playlist).Association("Songs").Delete(&song)
	if err != nil {
		return err
	}

	// Force update UpdatedAt
	ps.db.Model(&playlist).Update("updated_at", time.Now())

	return nil
}

func (ps *PlaylistStore) DeletePlaylist(playlistID uuid.UUID, userID uuid.UUID, adminOverride bool) error {
	if adminOverride {
		return ps.db.Delete(&model.Playlist{}, "id = ?", playlistID).Error
	}
	return ps.db.Delete(&model.Playlist{}, "id = ? AND user_id = ?", playlistID, userID).Error
}

func (ps *PlaylistStore) RenamePlaylist(playlistID uuid.UUID, userID uuid.UUID, newName string, adminOverride bool) error {
	if adminOverride {
		return ps.db.Model(&model.Playlist{}).Where("id = ?", playlistID).Update("name", newName).Error
	}
	return ps.db.Model(&model.Playlist{}).Where("id = ? AND user_id = ?", playlistID, userID).Update("name", newName).Error
}

func (ps *PlaylistStore) MovePlaylistToFolder(playlistID uuid.UUID, targetFolderID uuid.UUID, userID uuid.UUID, adminOverride bool) error {
	if adminOverride {
		return ps.db.Model(&model.Playlist{}).Where("id = ?", playlistID).Update("folder_id", targetFolderID).Error
	}
	return ps.db.Model(&model.Playlist{}).Where("id = ? AND user_id = ?", playlistID, userID).Update("folder_id", targetFolderID).Error
}

func (ps *PlaylistStore) RenamePlaylistFolder(folderID uuid.UUID, userID uuid.UUID, newName string, adminOverride bool) error {
	if adminOverride {
		return ps.db.Model(&model.PlaylistFolder{}).Where("id = ?", folderID).Update("name", newName).Error
	}
	return ps.db.Model(&model.PlaylistFolder{}).Where("id = ? AND user_id = ?", folderID, userID).Update("name", newName).Error
}

func (ps *PlaylistStore) MoveFolderToFolder(folderID uuid.UUID, targetParentID *uuid.UUID, userID uuid.UUID, adminOverride bool) error {
	if adminOverride {
		return ps.db.Model(&model.PlaylistFolder{}).Where("id = ?", folderID).Update("parent_id", targetParentID).Error
	}
	return ps.db.Model(&model.PlaylistFolder{}).Where("id = ? AND user_id = ?", folderID, userID).Update("parent_id", targetParentID).Error
}
func (ps *PlaylistStore) GetAllPlaylists() ([]model.Playlist, error) {
	var playlists []model.Playlist
	if err := ps.db.Find(&playlists).Error; err != nil {
		return nil, err
	}
	return playlists, nil
}

func (ps *PlaylistStore) GetChangedPlaylists(userID uuid.UUID, since time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := ps.db.Model(&model.Playlist{}).
		Where("user_id = ? AND updated_at > ?", userID, since).
		Pluck("id", &ids).Error
	return ids, err
}

func (ps *PlaylistStore) GetDeletedPlaylists(userID uuid.UUID, since time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := ps.db.Model(&model.Playlist{}).
		Unscoped().
		Where("user_id = ? AND deleted_at > ?", userID, since).
		Pluck("id", &ids).Error
	return ids, err
}

func (ps *PlaylistStore) GetChangedFolders(userID uuid.UUID, since time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := ps.db.Model(&model.PlaylistFolder{}).
		Where("user_id = ? AND updated_at > ?", userID, since).
		Pluck("id", &ids).Error
	return ids, err
}

func (ps *PlaylistStore) GetDeletedFolders(userID uuid.UUID, since time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := ps.db.Model(&model.PlaylistFolder{}).
		Unscoped().
		Where("user_id = ? AND deleted_at > ?", userID, since).
		Pluck("id", &ids).Error
	return ids, err
}

func (ps *PlaylistStore) GetPlaylistsByIDs(ids []uuid.UUID) ([]model.Playlist, error) {
	var playlists []model.Playlist
	err := ps.db.Preload("Songs").Where("id IN ?", ids).Find(&playlists).Error
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

func (ps *PlaylistStore) GetFoldersByIDs(ids []uuid.UUID) ([]model.PlaylistFolder, error) {
	var folders []model.PlaylistFolder
	err := ps.db.Where("id IN ?", ids).Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func (ps *PlaylistStore) GetRootFolderID(userID uuid.UUID) (uuid.UUID, error) {
	var folder model.PlaylistFolder
	err := ps.db.Where("user_id = ? AND parent_id IS NULL", userID).First(&folder).Error
	if err != nil {
		return uuid.Nil, err
	}
	return folder.ID, nil
}

func (ps *PlaylistStore) GetPlaylistsPaginated(page, limit int) ([]model.Playlist, bool, error) {
	return Paginate[model.Playlist](ps.db, page, limit, "name asc", []string{"Songs", "PlaylistFolder.User"})
}
