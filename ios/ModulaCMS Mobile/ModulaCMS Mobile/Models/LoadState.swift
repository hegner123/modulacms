enum LoadState<T> {
    case idle
    case loading
    case loaded(T)
    case error(String)
}
