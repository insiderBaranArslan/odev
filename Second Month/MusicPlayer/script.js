// Access playlist and control buttons
const playlistSongs = document.getElementById("playlist-songs");
const playButton = document.getElementById("play");
const pauseButton = document.getElementById("pause");
const nextButton = document.getElementById("next");
const previousButton = document.getElementById("previous");
const shuffleButton = document.getElementById("shuffle");

// Song list
const allSongs = [
  {
    id: 0,
    title: "Scratching The Surface",
    artist: "Quincy Larson",
    duration: "4:25",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/scratching-the-surface.mp3",
  },
  {
    id: 1,
    title: "Can't Stay Down",
    artist: "Quincy Larson",
    duration: "4:15",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/cant-stay-down.mp3",
  },
  {
    id: 2,
    title: "Still Learning",
    artist: "Quincy Larson",
    duration: "3:51",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/still-learning.mp3",
  },
  {
    id: 3,
    title: "Cruising for a Musing",
    artist: "Quincy Larson",
    duration: "3:34",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/cruising-for-a-musing.mp3",
  },
  {
    id: 4,
    title: "Never Not Favored",
    artist: "Quincy Larson",
    duration: "3:35",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/never-not-favored.mp3",
  },
  {
    id: 5,
    title: "From the Ground Up",
    artist: "Quincy Larson",
    duration: "3:12",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/from-the-ground-up.mp3",
  },
  {
    id: 6,
    title: "Walking on Air",
    artist: "Quincy Larson",
    duration: "3:25",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/walking-on-air.mp3",
  },
  {
    id: 7,
    title: "Can't Stop Me. Can't Even Slow Me Down.",
    artist: "Quincy Larson",
    duration: "3:52",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/cant-stop-me-cant-even-slow-me-down.mp3",
  },
  {
    id: 8,
    title: "The Surest Way Out",
    artist: "Quincy Larson",
    duration: "3:18",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/the-surest-way-out.mp3",
  },
  {
    id: 9,
    title: "Chasing That Feeling",
    artist: "Quincy Larson",
    duration: "2:43",
    src: "https://cdn.freecodecamp.org/curriculum/js-music-player/chasing-that-feeling.mp3",
  },
];

// Audio object
const audio = new Audio();

// App state
let userData = {
  songs: [...allSongs],
  currentSong: null,
  songCurrentTime: 0,
};

// Play a song
const playSong = (id) => {
  const song = userData.songs.find((s) => s.id === id);
  audio.src = song.src;
  audio.title = song.title;

  if (userData.currentSong === null || userData.currentSong.id !== song.id) {
    audio.currentTime = 0;
  } else {
    audio.currentTime = userData.songCurrentTime;
  }

  userData.currentSong = song;
  playButton.classList.add("playing");

  highlightCurrentSong();
  setPlayerDisplay();
  setPlayButtonAccessibleText();
  audio.play();
};

// Pause
const pauseSong = () => {
  userData.songCurrentTime = audio.currentTime;
  playButton.classList.remove("playing");
  audio.pause();
};

// Next song
const playNextSong = () => {
  if (userData.currentSong === null) {
    playSong(userData.songs[0].id);
  } else {
    const currentSongIndex = getCurrentSongIndex();
    const nextSong = userData.songs[currentSongIndex + 1];
    playSong(nextSong ? nextSong.id : userData.songs[0].id);
  }
};

// Previous song
const playPreviousSong = () => {
  if (userData.currentSong === null) return;
  const currentSongIndex = getCurrentSongIndex();
  const previousSong = userData.songs[currentSongIndex - 1];
  playSong(previousSong ? previousSong.id : userData.songs[userData.songs.length - 1].id);
};

// Shuffle
const shuffle = () => {
  userData.songs = userData.songs
    .map((song) => ({ song, sort: Math.random() }))
    .sort((a, b) => a.sort - b.sort)
    .map(({ song }) => song);

  userData.currentSong = null;
  userData.songCurrentTime = 0;

  renderSongs(userData.songs);
  pauseSong();
  setPlayerDisplay();
  setPlayButtonAccessibleText();
};

// Delete a song
const deleteSong = (id) => {
  if (userData.currentSong && userData.currentSong.id === id) {
    userData.currentSong = null;
    userData.songCurrentTime = 0;
    pauseSong();
    setPlayerDisplay();
  }

  userData.songs = userData.songs.filter((s) => s.id !== id);
  renderSongs(userData.songs);
  highlightCurrentSong();
  setPlayButtonAccessibleText();

  if (userData.songs.length === 0) {
    const resetButton = document.createElement("button");
    const resetText = document.createTextNode("Reset Playlist");
    resetButton.id = "reset";
    resetButton.ariaLabel = "Reset playlist";
    resetButton.appendChild(resetText);
    playlistSongs.appendChild(resetButton);
    resetButton.addEventListener("click", () => {
      userData.songs = [...allSongs];
      renderSongs(userData.songs);
      setPlayButtonAccessibleText();
    });
  }
};

// Update player display
const setPlayerDisplay = () => {
  const playingSong = document.getElementById("player-song-title");
  const songArtist = document.getElementById("player-song-artist");
  const currentTitle = userData.currentSong?.title ?? "";
  const currentArtist = userData.currentSong?.artist ?? "";
  playingSong.textContent = currentTitle;
  songArtist.textContent = currentArtist;
};

// Highlight the current song in the playlist
const highlightCurrentSong = () => {
  const playlistSongElements = document.querySelectorAll(".playlist-song");
  const songToHighlight = document.getElementById(`song-${userData.currentSong?.id}`);

  playlistSongElements.forEach((el) => el.removeAttribute("aria-current"));
  if (songToHighlight) songToHighlight.setAttribute("aria-current", "true");
};

// Render songs to the DOM
const renderSongs = (array) => {
  const songsHTML = array
    .map((song) => {
      return `
        <li id="song-${song.id}" class="playlist-song">
          <button class="playlist-song-info" onclick="playSong(${song.id})">
            <span class="playlist-song-title">${song.title}</span>
            <span class="playlist-song-artist">${song.artist}</span>
            <span class="playlist-song-duration">${song.duration}</span>
          </button>
          <button class="playlist-song-delete" aria-label="Delete ${song.title}" onclick="deleteSong(${song.id})">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="8" cy="8" r="8" fill="#4d4d62"/>
              <path fill-rule="evenodd" clip-rule="evenodd" d="M4.054 4.054a.5.5 0 01.707 0L8 7.293l3.239-3.24a.5.5 0 11.707.708L8.707 8l3.24 3.239a.5.5 0 01-.707.707L8 8.707l-3.239 3.24a.5.5 0 01-.707-.707L7.293 8 4.054 4.761a.5.5 0 010-.707z" fill="#dfdfe2"/>
            </svg>
          </button>
        </li>
      `;
    })
    .join("");
  playlistSongs.innerHTML = songsHTML;
};

// Set accessible text for play button
const setPlayButtonAccessibleText = () => {
  const song = userData.currentSong ?? userData.songs[0];
  playButton.setAttribute(
    "aria-label",
    song?.title ? `Play ${song.title}` : "Play"
  );
};

// Get current song index
const getCurrentSongIndex = () =>
  userData.songs.indexOf(userData.currentSong);

// Button event listeners
playButton.addEventListener("click", () => {
  if (userData.currentSong === null) {
    playSong(userData.songs[0].id);
  } else {
    playSong(userData.currentSong.id);
  }
});

pauseButton.addEventListener("click", pauseSong);
nextButton.addEventListener("click", playNextSong);
previousButton.addEventListener("click", playPreviousSong);
shuffleButton.addEventListener("click", shuffle);

// Auto-play next song when current ends
audio.addEventListener("ended", () => {
  const nextIndex = getCurrentSongIndex() + 1;
  if (nextIndex < userData.songs.length) {
    playSong(userData.songs[nextIndex].id);
  } else {
    userData.currentSong = null;
    userData.songCurrentTime = 0;
    pauseSong();
    setPlayerDisplay();
    highlightCurrentSong();
    setPlayButtonAccessibleText();
  }
});

// Initial render
renderSongs(userData.songs);
setPlayButtonAccessibleText();
