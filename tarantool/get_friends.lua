function get_friends(profile_id)
  local result_set = {}
  local result_index = 0
  for _,tuple in box.space['friends']:pairs{profile_id} do
    for _,profile in box.space['profile']:pairs{tuple[2]} do
      result_index = result_index + 1
      result_set[result_index] = profile
    end
  end
  return result_set
end